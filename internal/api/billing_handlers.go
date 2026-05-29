package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ZaViBiS/isitdead/internal/billing"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v83"
	billingportal_session "github.com/stripe/stripe-go/v83/billingportal/session"
	checkout_session "github.com/stripe/stripe-go/v83/checkout/session"
	"github.com/stripe/stripe-go/v83/subscription"
	"github.com/stripe/stripe-go/v83/webhook"
)

func (s *Server) billingPriceIDs() billing.PriceIDs {
	return billing.PriceIDs{
		Pro:      strings.TrimSpace(s.Config.StripeProPriceID),
		Business: strings.TrimSpace(s.Config.StripeBusinessPriceID),
	}
}

func (s *Server) appBaseURL() string {
	if s.Config.Env == "dev" || s.Config.Domain == "localhost" {
		return fmt.Sprintf("http://localhost:%s", s.Config.Port)
	}
	return fmt.Sprintf("https://%s", strings.TrimSuffix(s.Config.Domain, "/"))
}

func (s *Server) handleGetBillingPlans(c fiber.Ctx) error {
	return c.JSON(billing.Plans(s.billingPriceIDs()))
}

func (s *Server) handleCreateCheckoutSession(c fiber.Ctx) error {
	if s.Config.StripeSecretKey == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Stripe is not configured"})
	}

	userID := c.Locals("user_id").(uint)
	user, err := s.DB.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}

	var req checkoutRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	plan := billing.PlanByID(req.Plan, s.billingPriceIDs())
	if plan.ID == billing.PlanFree || plan.StripePriceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "This plan cannot be purchased"})
	}

	stripe.Key = s.Config.StripeSecretKey
	if user.StripeSubscriptionID != "" {
		if user.StripePriceID == plan.StripePriceID {
			return c.JSON(checkoutResponse{URL: "/pricing"})
		}
		if err := s.updateStripeSubscriptionPlan(user.ID, user.StripeSubscriptionID, plan.ID, plan.StripePriceID); err != nil {
			log.Error().Err(err).Uint("user_id", user.ID).Str("plan", plan.ID).Msg("failed to update Stripe subscription")
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Could not update subscription"})
		}
		return c.JSON(checkoutResponse{URL: "/pricing?checkout=success"})
	}

	params := &stripe.CheckoutSessionParams{
		SuccessURL:        stripe.String(s.appBaseURL() + "/pricing?checkout=success"),
		CancelURL:         stripe.String(s.appBaseURL() + "/pricing?checkout=cancelled"),
		Mode:              stripe.String(stripe.CheckoutSessionModeSubscription),
		ClientReferenceID: stripe.String(strconv.FormatUint(uint64(user.ID), 10)),
		CustomerEmail:     stripe.String(user.Email),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(plan.StripePriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"user_id": strconv.FormatUint(uint64(user.ID), 10),
			"plan":    plan.ID,
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"user_id": strconv.FormatUint(uint64(user.ID), 10),
				"plan":    plan.ID,
			},
		},
	}
	if user.StripeCustomerID != "" {
		params.Customer = stripe.String(user.StripeCustomerID)
		params.CustomerEmail = nil
	}

	session, err := checkout_session.New(params)
	if err != nil {
		log.Error().Err(err).Uint("user_id", user.ID).Str("plan", plan.ID).Msg("failed to create Stripe checkout session")
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Could not start checkout"})
	}

	return c.JSON(checkoutResponse{URL: session.URL})
}

func (s *Server) updateStripeSubscriptionPlan(userID uint, subscriptionID, planID, priceID string) error {
	current, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return err
	}
	if current.Items == nil || len(current.Items.Data) == 0 || current.Items.Data[0] == nil {
		return fmt.Errorf("Stripe subscription %s has no subscription items", subscriptionID)
	}

	updated, err := subscription.Update(subscriptionID, &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:       stripe.String(current.Items.Data[0].ID),
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"user_id": strconv.FormatUint(uint64(userID), 10),
			"plan":    planID,
		},
		ProrationBehavior: stripe.String("create_prorations"),
	})
	if err != nil {
		return err
	}

	customerID := ""
	if updated.Customer != nil {
		customerID = updated.Customer.ID
	}
	return s.DB.UpdateUserBilling(
		userID,
		planID,
		customerID,
		updated.ID,
		string(updated.Status),
		subscriptionPriceID(*updated),
		subscriptionPeriodEnd(*updated),
	)
}

func (s *Server) handleCreateBillingPortalSession(c fiber.Ctx) error {
	if s.Config.StripeSecretKey == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Stripe is not configured"})
	}

	userID := c.Locals("user_id").(uint)
	user, err := s.DB.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user"})
	}
	if user.StripeCustomerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No Stripe customer is linked to this account"})
	}

	stripe.Key = s.Config.StripeSecretKey
	session, err := billingportal_session.New(&stripe.BillingPortalSessionParams{
		Customer:  stripe.String(user.StripeCustomerID),
		ReturnURL: stripe.String(s.appBaseURL() + "/pricing"),
	})
	if err != nil {
		log.Error().Err(err).Uint("user_id", user.ID).Msg("failed to create Stripe billing portal session")
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Could not open billing portal"})
	}

	return c.JSON(checkoutResponse{URL: session.URL})
}

func (s *Server) handleStripeWebhook(c fiber.Ctx) error {
	if s.Config.StripeWebhookSecret == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Stripe webhook is not configured"})
	}

	event, err := webhook.ConstructEventWithOptions(
		c.Body(),
		c.Get("Stripe-Signature"),
		s.Config.StripeWebhookSecret,
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true},
	)
	if err != nil {
		log.Warn().Err(err).Msg("invalid Stripe webhook")
		return c.SendStatus(fiber.StatusBadRequest)
	}

	switch event.Type {
	case "checkout.session.completed":
		if err := s.handleCheckoutCompleted(event); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not process checkout"})
		}
	case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted":
		if err := s.handleSubscriptionChanged(event); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not process subscription"})
		}
	default:
		log.Debug().Str("event_type", string(event.Type)).Msg("ignored Stripe webhook event")
	}

	return c.SendStatus(fiber.StatusOK)
}

func (s *Server) handleCheckoutCompleted(event stripe.Event) error {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return err
	}

	userID, err := strconv.ParseUint(session.Metadata["user_id"], 10, 64)
	if err != nil {
		return err
	}

	planID := session.Metadata["plan"]
	if planID == "" && session.Subscription != nil {
		planID = session.Subscription.Metadata["plan"]
	}
	if !billing.IsPaidPlan(planID) {
		planID = billing.PlanFree
	}

	customerID := ""
	if session.Customer != nil {
		customerID = session.Customer.ID
	}
	subscriptionID := ""
	if session.Subscription != nil {
		subscriptionID = session.Subscription.ID
	}

	return s.DB.UpdateUserBilling(uint(userID), planID, customerID, subscriptionID, "active", "", nil)
}

func (s *Server) handleSubscriptionChanged(event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return err
	}

	customerID := ""
	if subscription.Customer != nil {
		customerID = subscription.Customer.ID
	}

	user, err := s.DB.GetUserByStripeSubscriptionID(subscription.ID)
	if err != nil && customerID != "" {
		user, err = s.DB.GetUserByStripeCustomerID(customerID)
	}
	if err != nil && subscription.Metadata["user_id"] != "" {
		if userID, parseErr := strconv.ParseUint(subscription.Metadata["user_id"], 10, 64); parseErr == nil {
			user, err = s.DB.GetUserByID(uint(userID))
		}
	}
	if err != nil {
		return err
	}
	if user.StripeSubscriptionID != "" && user.StripeSubscriptionID != subscription.ID {
		log.Info().
			Uint("user_id", user.ID).
			Str("current_subscription_id", user.StripeSubscriptionID).
			Str("event_subscription_id", subscription.ID).
			Str("event_status", string(subscription.Status)).
			Msg("ignored Stripe event for non-current subscription")
		return nil
	}

	priceID := subscriptionPriceID(subscription)
	planID := billing.PlanByPriceID(priceID, s.billingPriceIDs())
	if !subscriptionAllowsAccess(subscription.Status) {
		planID = billing.PlanFree
	}

	return s.DB.UpdateUserBilling(
		user.ID,
		planID,
		customerID,
		subscription.ID,
		string(subscription.Status),
		priceID,
		subscriptionPeriodEnd(subscription),
	)
}

func subscriptionAllowsAccess(status stripe.SubscriptionStatus) bool {
	return status == stripe.SubscriptionStatusActive || status == stripe.SubscriptionStatusTrialing
}

func subscriptionPriceID(subscription stripe.Subscription) string {
	if subscription.Items == nil || len(subscription.Items.Data) == 0 || subscription.Items.Data[0].Price == nil {
		return ""
	}
	return subscription.Items.Data[0].Price.ID
}

func subscriptionPeriodEnd(subscription stripe.Subscription) *time.Time {
	if subscription.Items == nil || len(subscription.Items.Data) == 0 {
		return nil
	}
	periodEnd := subscription.Items.Data[0].CurrentPeriodEnd
	if periodEnd == 0 {
		return nil
	}
	t := time.Unix(periodEnd, 0).UTC()
	return &t
}
