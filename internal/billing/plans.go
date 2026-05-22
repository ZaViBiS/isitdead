package billing

const (
	PlanFree     = "free"
	PlanPro      = "pro"
	PlanBusiness = "business"
)

type Plan struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Price           string `json:"price"`
	StripePriceID   string `json:"-"`
	MonitorLimit    int    `json:"monitor_limit"`
	MinInterval     int    `json:"min_interval"`
	HistoryDays     int    `json:"history_days"`
	PublicPages     bool   `json:"public_pages"`
	SSLMonitoring   bool   `json:"ssl_monitoring"`
	TelegramAlerts  bool   `json:"telegram_alerts"`
	StripeAvailable bool   `json:"stripe_available"`
}

type PriceIDs struct {
	Pro      string
	Business string
}

func Plans(priceIDs PriceIDs) []Plan {
	plans := []Plan{
		{
			ID:             PlanFree,
			Name:           "Free",
			Description:    "For personal projects and small checks.",
			Price:          "$0",
			MonitorLimit:   3,
			MinInterval:    300,
			HistoryDays:    7,
			PublicPages:    false,
			SSLMonitoring:  true,
			TelegramAlerts: false,
		},
		{
			ID:             PlanPro,
			Name:           "Pro",
			Description:    "For teams that need faster checks and more monitors.",
			Price:          "$9/mo",
			StripePriceID:  priceIDs.Pro,
			MonitorLimit:   25,
			MinInterval:    60,
			HistoryDays:    30,
			PublicPages:    true,
			SSLMonitoring:  true,
			TelegramAlerts: true,
		},
		{
			ID:             PlanBusiness,
			Name:           "Business",
			Description:    "For production services with wider monitoring needs.",
			Price:          "$29/mo",
			StripePriceID:  priceIDs.Business,
			MonitorLimit:   100,
			MinInterval:    30,
			HistoryDays:    90,
			PublicPages:    true,
			SSLMonitoring:  true,
			TelegramAlerts: true,
		},
	}

	for i := range plans {
		plans[i].StripeAvailable = plans[i].ID == PlanFree || plans[i].StripePriceID != ""
	}
	return plans
}

func PlanByID(planID string, priceIDs PriceIDs) Plan {
	for _, plan := range Plans(priceIDs) {
		if plan.ID == planID {
			return plan
		}
	}
	return Plans(priceIDs)[0]
}

func PlanByPriceID(priceID string, priceIDs PriceIDs) string {
	switch priceID {
	case priceIDs.Pro:
		return PlanPro
	case priceIDs.Business:
		return PlanBusiness
	default:
		return PlanFree
	}
}

func IsPaidPlan(planID string) bool {
	return planID == PlanPro || planID == PlanBusiness
}
