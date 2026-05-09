// Package logger надає функціонал для логування подій.
package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Setup налаштовує глобальний логер zerolog
func Setup(debug bool) {
	// Встановлюємо формат часу
	zerolog.TimeFieldFormat = time.RFC3339

	// Рівень логування: Debug для розробки, Info для продакшну
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		// Для розробки використовуємо ConsoleWriter (гарний кольоровий вивід)
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
		})
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		// Для продакшну залишаємо JSON формат (за замовчуванням у zerolog)
		// Він краще підходить для збору логів (ELK, Loki тощо)
	}

	log.Info().Bool("debug", debug).Msg("Logger initialized")
}
