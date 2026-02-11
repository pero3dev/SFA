package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	dbgen "sfa/backend/internal/db/sqlc"
	"sfa/backend/internal/store"
)

type DashboardHandler struct {
	Store *store.Store
}

func NewDashboardHandler(store *store.Store) DashboardHandler {
	return DashboardHandler{Store: store}
}

func (h DashboardHandler) KPI(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]string{
				"code":    "invalid_tenant_id",
				"message": err.Error(),
			},
		})
		return
	}

	var rows []dbgen.GetLatestKpiSnapshotRow
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		var queryErr error
		rows, queryErr = q.GetLatestKpiSnapshot(r.Context(), toPGUUID(tenantID))
		return queryErr
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": map[string]string{
				"code":    "kpi_query_failed",
				"message": "failed to fetch kpi snapshot",
			},
		})
		return
	}

	snapshotAt := time.Now().UTC()
	if len(rows) > 0 && rows[0].SnapshotAt.Valid {
		snapshotAt = rows[0].SnapshotAt.Time.UTC()
	}

	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		dimensions := map[string]any{}
		if len(row.Dimensions) > 0 {
			if err := json.Unmarshal(row.Dimensions, &dimensions); err != nil {
				dimensions = map[string]any{}
			}
		}

		items = append(items, map[string]any{
			"metricKey":   row.MetricKey,
			"metricValue": row.MetricValue,
			"dimensions":  dimensions,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"snapshotAt": snapshotAt.Format(time.RFC3339),
		"data":       items,
	})
}

func (h DashboardHandler) Pipeline(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]string{
				"code":    "invalid_tenant_id",
				"message": err.Error(),
			},
		})
		return
	}

	var rows []dbgen.GetPipelineSummaryRow
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		var queryErr error
		rows, queryErr = q.GetPipelineSummary(r.Context(), toPGUUID(tenantID))
		return queryErr
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error": map[string]string{
				"code":    "pipeline_query_failed",
				"message": "failed to fetch pipeline summary",
			},
		})
		return
	}

	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, map[string]any{
			"stage":       string(row.Stage),
			"count":       row.DealCount,
			"totalAmount": row.TotalAmount,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": items,
	})
}
