package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"sfa/backend/internal/http/handlers"
	"sfa/backend/internal/store"
)

func NewRouter(store *store.Store) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	healthHandler := handlers.NewHealthHandler(store)
	r.Get("/livez", healthHandler.Live)
	r.Get("/readyz", healthHandler.Ready)

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/health", healthHandler.Live)

		// Endpoint placeholders aligned with api/openapi.yaml
		registerAuthRoutes(api)
		registerUserRoutes(api)
		registerAccountRoutes(api)
		registerOpportunityRoutes(api)
		registerDashboardRoutes(api, store)
		registerAuditRoutes(api)
		registerFeaturePackRoutes(api, store)
	})

	return r
}

func registerAuthRoutes(r chi.Router) {
	r.Route("/auth", func(auth chi.Router) {
		auth.Post("/login", notImplemented)
		auth.Post("/refresh", notImplemented)
		auth.Post("/logout", notImplemented)
		auth.Get("/me", notImplemented)
	})
}

func registerUserRoutes(r chi.Router) {
	r.Route("/users", func(users chi.Router) {
		users.Get("/", notImplemented)
		users.Post("/", notImplemented)
		users.Patch("/{id}", notImplemented)
	})
}

func registerAccountRoutes(r chi.Router) {
	r.Route("/accounts", func(accounts chi.Router) {
		accounts.Get("/", notImplemented)
		accounts.Post("/", notImplemented)
		accounts.Patch("/{id}", notImplemented)

		accounts.Route("/{id}/contacts", func(contacts chi.Router) {
			contacts.Get("/", notImplemented)
			contacts.Post("/", notImplemented)
		})

		accounts.Route("/{id}/locations", func(locations chi.Router) {
			locations.Get("/", notImplemented)
			locations.Post("/", notImplemented)
		})
	})
}

func registerOpportunityRoutes(r chi.Router) {
	r.Route("/opportunities", func(opps chi.Router) {
		opps.Get("/", notImplemented)
		opps.Post("/", notImplemented)
		opps.Patch("/{id}", notImplemented)

		opps.Route("/{id}/activities", func(activities chi.Router) {
			activities.Get("/", notImplemented)
			activities.Post("/", notImplemented)
		})
		opps.Route("/{id}/quotes", func(quotes chi.Router) {
			quotes.Get("/", notImplemented)
			quotes.Post("/", notImplemented)
		})
		opps.Route("/{id}/orders", func(orders chi.Router) {
			orders.Get("/", notImplemented)
			orders.Post("/", notImplemented)
		})
		opps.Post("/{id}/lost", notImplemented)
	})
}

func registerDashboardRoutes(r chi.Router, store *store.Store) {
	dashboardHandler := handlers.NewDashboardHandler(store)

	r.Route("/dashboard", func(d chi.Router) {
		d.Get("/kpi", dashboardHandler.KPI)
		d.Get("/pipeline", dashboardHandler.Pipeline)
	})
}

func registerAuditRoutes(r chi.Router) {
	r.Get("/audit-logs", notImplemented)
}

func registerFeaturePackRoutes(r chi.Router, store *store.Store) {
	features := handlers.NewFeaturePackHandler(store)

	r.Get("/opportunities/next-actions", features.NextActions)
	r.Patch("/opportunities/{id}/next-action", features.UpdateNextAction)

	r.Route("/analytics", func(analytics chi.Router) {
		analytics.Get("/deal-health", features.DealHealth)
		analytics.Get("/forecast", features.Forecast)
		analytics.Get("/loss-reasons", features.LossReasonAnalysis)
		analytics.Get("/duplicates", features.DuplicateCandidates)
	})

	r.Route("/integrations", func(integrations chi.Router) {
		integrations.Get("/connections", features.ListIntegrationConnections)
		integrations.Post("/connections", features.UpsertIntegrationConnection)
		integrations.Get("/events", features.ListIntegrationEvents)
		integrations.Post("/events", features.CreateIntegrationEvent)
	})

	r.Route("/approvals", func(approvals chi.Router) {
		approvals.Get("/", features.ListApprovalRequests)
		approvals.Post("/", features.CreateApprovalRequest)
		approvals.Post("/{id}/decision", features.DecideApproval)
	})

	r.Get("/export/accounts.csv", features.ExportAccountsCSV)
	r.Get("/export/opportunities.csv", features.ExportOpportunitiesCSV)
	r.Post("/import/accounts.csv", features.ImportAccountsCSV)
	r.Post("/import/opportunities.csv", features.ImportOpportunitiesCSV)
}

func notImplemented(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(`{"error":{"code":"not_implemented","message":"endpoint scaffolded"}}`))
}
