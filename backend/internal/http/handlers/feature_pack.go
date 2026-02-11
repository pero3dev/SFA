package handlers

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "sfa/backend/internal/db/sqlc"
	"sfa/backend/internal/store"
)

type FeaturePackHandler struct {
	Store *store.Store
}

func NewFeaturePackHandler(store *store.Store) FeaturePackHandler {
	return FeaturePackHandler{Store: store}
}

func (h FeaturePackHandler) NextActions(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error())
		return
	}

	offset, limit := queryPageLimit(r, 20)
	dueBefore, err := parseOptionalTimestamp(r.URL.Query().Get("dueBefore"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_due_before", "dueBefore must be RFC3339 or YYYY-MM-DD")
		return
	}

	var rows []dbgen.ListNextActionsRow
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		var queryErr error
		rows, queryErr = q.ListNextActions(r.Context(), dbgen.ListNextActionsParams{
			TenantID:    toPGUUID(tenantID),
			DueBefore:   dueBefore,
			OffsetCount: offset,
			LimitCount:  limit,
		})
		return queryErr
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "next_actions_failed", "failed to load next actions")
		return
	}

	data := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		data = append(data, map[string]any{
			"id":             pgUUIDToString(row.ID),
			"name":           row.Name,
			"stage":          string(row.Stage),
			"accountName":    row.AccountName,
			"nextActionAt":   pgTimestampToString(row.NextActionAt),
			"nextActionNote": pgTextToString(row.NextActionNote),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": data,
		"meta": map[string]any{
			"limit": limit,
		},
	})
}

func (h FeaturePackHandler) UpdateNextAction(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error())
		return
	}

	opportunityID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_opportunity_id", "id must be UUID")
		return
	}

	var req struct {
		NextActionAt   string `json:"nextActionAt"`
		NextActionNote string `json:"nextActionNote"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}

	actionAt, err := time.Parse(time.RFC3339, req.NextActionAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_next_action_at", "nextActionAt must be RFC3339")
		return
	}

	var row dbgen.Opportunity
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		var queryErr error
		row, queryErr = q.UpdateOpportunityNextAction(r.Context(), dbgen.UpdateOpportunityNextActionParams{
			NextActionAt:   toPGTimestamptz(actionAt.UTC()),
			NextActionNote: toPGText(req.NextActionNote),
			TenantID:       toPGUUID(tenantID),
			OpportunityID:  toPGUUID(opportunityID),
		})
		return queryErr
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "opportunity not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "update_next_action_failed", "failed to update next action")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":             pgUUIDToString(row.ID),
			"name":           row.Name,
			"nextActionAt":   pgTimestampToString(row.NextActionAt),
			"nextActionNote": pgTextToString(row.NextActionNote),
		},
	})
}

func (h FeaturePackHandler) DealHealth(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error())
		return
	}

	offset, limit := queryPageLimit(r, 20)
	var rows []dbgen.ListDealHealthRow
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		var queryErr error
		rows, queryErr = q.ListDealHealth(r.Context(), dbgen.ListDealHealthParams{
			TenantID:    toPGUUID(tenantID),
			OffsetCount: offset,
			LimitCount:  limit,
		})
		return queryErr
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "deal_health_failed", "failed to load deal health")
		return
	}

	data := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		data = append(data, map[string]any{
			"id":             pgUUIDToString(row.ID),
			"name":           row.Name,
			"stage":          string(row.Stage),
			"probability":    row.Probability,
			"amount":         row.Amount,
			"lastActivityAt": pgTimestampToString(row.LastActivityAt),
			"healthScore":    row.HealthScore,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func (h FeaturePackHandler) Forecast(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error())
		return
	}

	var rows []dbgen.GetForecastSummaryRow
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		var queryErr error
		rows, queryErr = q.GetForecastSummary(r.Context(), toPGUUID(tenantID))
		return queryErr
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "forecast_failed", "failed to load forecast")
		return
	}

	data := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		data = append(data, map[string]any{
			"ownerUserId":    pgUUIDToString(row.OwnerUserID),
			"month":          row.MonthBucket,
			"dealCount":      row.DealCount,
			"pipelineAmount": row.PipelineAmount,
			"weightedAmount": row.WeightedAmount,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func (h FeaturePackHandler) LossReasonAnalysis(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error())
		return
	}

	var rows []dbgen.GetLossReasonAnalysisRow
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		var queryErr error
		rows, queryErr = q.GetLossReasonAnalysis(r.Context(), toPGUUID(tenantID))
		return queryErr
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "loss_analysis_failed", "failed to load loss reason analysis")
		return
	}

	data := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		data = append(data, map[string]any{
			"reason":     string(row.Reason),
			"lostCount":  row.LostCount,
			"lostAmount": row.LostAmount,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func (h FeaturePackHandler) DuplicateCandidates(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error())
		return
	}

	var rows []dbgen.ListDuplicateCandidatesRow
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		var queryErr error
		rows, queryErr = q.ListDuplicateCandidates(r.Context(), toPGUUID(tenantID))
		return queryErr
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "duplicate_scan_failed", "failed to detect duplicates")
		return
	}

	data := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		data = append(data, map[string]any{
			"type":        row.DuplicateType,
			"primaryId":   pgUUIDToString(row.PrimaryID),
			"duplicateId": pgUUIDToString(row.DuplicateID),
			"matchValue":  row.MatchValue,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func (h FeaturePackHandler) UpsertIntegrationConnection(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error())
		return
	}

	var req struct {
		UserID            string   `json:"userId"`
		Provider          string   `json:"provider"`
		IntegrationType   string   `json:"integrationType"`
		ExternalAccountID string   `json:"externalAccountId"`
		Status            string   `json:"status"`
		AccessToken       string   `json:"accessToken"`
		RefreshToken      string   `json:"refreshToken"`
		ExpiresAt         string   `json:"expiresAt"`
		Scopes            []string `json:"scopes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}

	userID, err := parseUUID(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_user_id", "userId must be UUID")
		return
	}

	provider, err := parseProvider(req.Provider)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_provider", err.Error())
		return
	}
	integrationType, err := parseIntegrationType(req.IntegrationType)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_integration_type", err.Error())
		return
	}

	status := dbgen.NullIntegrationStatusEnum{}
	if req.Status != "" {
		if parsed, parseErr := parseIntegrationStatus(req.Status); parseErr == nil {
			status = dbgen.NullIntegrationStatusEnum{IntegrationStatusEnum: parsed, Valid: true}
		} else {
			writeError(w, http.StatusBadRequest, "invalid_status", parseErr.Error())
			return
		}
	}

	expiresAt, err := parseOptionalTimestamp(req.ExpiresAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_expires_at", "expiresAt must be RFC3339 or YYYY-MM-DD")
		return
	}

	var row dbgen.IntegrationConnection
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		var queryErr error
		row, queryErr = q.UpsertIntegrationConnection(r.Context(), dbgen.UpsertIntegrationConnectionParams{
			TenantID:          toPGUUID(tenantID),
			UserID:            toPGUUID(userID),
			Provider:          provider,
			IntegrationType:   integrationType,
			ExternalAccountID: req.ExternalAccountID,
			Status:            status,
			AccessToken:       toPGText(req.AccessToken),
			RefreshToken:      toPGText(req.RefreshToken),
			ExpiresAt:         expiresAt,
			Scopes:            req.Scopes,
		})
		return queryErr
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "integration_upsert_failed", "failed to save integration")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": integrationConnectionDTO(row)})
}

func (h FeaturePackHandler) ListIntegrationConnections(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error())
		return
	}

	var rows []dbgen.IntegrationConnection
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		var queryErr error
		rows, queryErr = q.ListIntegrationConnections(r.Context(), toPGUUID(tenantID))
		return queryErr
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "integration_list_failed", "failed to load integrations")
		return
	}

	data := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		data = append(data, integrationConnectionDTO(row))
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func (h FeaturePackHandler) CreateApprovalRequest(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error())
		return
	}

	var req struct {
		EntityType     string `json:"entityType"`
		EntityID       string `json:"entityId"`
		RequestedBy    string `json:"requestedBy"`
		ApproverUserID string `json:"approverUserId"`
		Reason         string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}

	entityID, err := parseUUID(req.EntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_entity_id", "entityId must be UUID")
		return
	}
	requestedBy, err := parseUUID(req.RequestedBy)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_requested_by", "requestedBy must be UUID")
		return
	}
	approverID, err := parseUUID(req.ApproverUserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_approver", "approverUserId must be UUID")
		return
	}

	var row dbgen.ApprovalRequest
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		var queryErr error
		row, queryErr = q.CreateApprovalRequest(r.Context(), dbgen.CreateApprovalRequestParams{
			TenantID:       toPGUUID(tenantID),
			EntityType:     req.EntityType,
			EntityID:       toPGUUID(entityID),
			RequestedBy:    toPGUUID(requestedBy),
			ApproverUserID: toPGUUID(approverID),
			Reason:         req.Reason,
		})
		return queryErr
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "approval_create_failed", "failed to create approval request")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": approvalDTO(row)})
}

func (h FeaturePackHandler) ListApprovalRequests(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error())
		return
	}

	offset, limit := queryPageLimit(r, 20)
	status := dbgen.NullApprovalStatusEnum{}
	if raw := r.URL.Query().Get("status"); raw != "" {
		parsed, parseErr := parseApprovalStatus(raw)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "invalid_status", parseErr.Error())
			return
		}
		status = dbgen.NullApprovalStatusEnum{ApprovalStatusEnum: parsed, Valid: true}
	}

	var rows []dbgen.ApprovalRequest
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		var queryErr error
		rows, queryErr = q.ListApprovalRequests(r.Context(), dbgen.ListApprovalRequestsParams{
			TenantID:    toPGUUID(tenantID),
			Status:      status,
			OffsetCount: offset,
			LimitCount:  limit,
		})
		return queryErr
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "approval_list_failed", "failed to load approvals")
		return
	}

	data := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		data = append(data, approvalDTO(row))
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func (h FeaturePackHandler) DecideApproval(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromHeader(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tenant_id", err.Error())
		return
	}

	approvalID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_approval_id", "id must be UUID")
		return
	}

	var req struct {
		Status       string `json:"status"`
		DecisionNote string `json:"decisionNote"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}
	status, err := parseApprovalStatus(req.Status)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_status", err.Error())
		return
	}

	var row dbgen.ApprovalRequest
	if err := h.Store.WithTenantTx(r.Context(), tenantID, func(q *dbgen.Queries) error {
		var queryErr error
		row, queryErr = q.DecideApprovalRequest(r.Context(), dbgen.DecideApprovalRequestParams{
			Status:       status,
			DecisionNote: toPGText(req.DecisionNote),
			TenantID:     toPGUUID(tenantID),
			ApprovalID:   toPGUUID(approvalID),
		})
		return queryErr
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "approval request not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "approval_decision_failed", "failed to decide approval")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": approvalDTO(row)})
}

func (h FeaturePackHandler) ExportAccountsCSV(w http.ResponseWriter, r *http.Request) {
	h.exportAccountsCSV(w, r)
}

func (h FeaturePackHandler) ExportOpportunitiesCSV(w http.ResponseWriter, r *http.Request) {
	h.exportOpportunitiesCSV(w, r)
}

func (h FeaturePackHandler) ImportAccountsCSV(w http.ResponseWriter, r *http.Request) {
	h.importAccountsCSV(w, r)
}

func (h FeaturePackHandler) ImportOpportunitiesCSV(w http.ResponseWriter, r *http.Request) {
	h.importOpportunitiesCSV(w, r)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func parseOptionalTimestamp(raw string) (pgtype.Timestamptz, error) {
	if strings.TrimSpace(raw) == "" {
		return pgtype.Timestamptz{}, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return toPGTimestamptz(t.UTC()), nil
	}
	if d, err := time.Parse("2006-01-02", raw); err == nil {
		return toPGTimestamptz(d.UTC()), nil
	}
	return pgtype.Timestamptz{}, errors.New("invalid timestamp")
}

func parseProvider(raw string) (dbgen.IntegrationProviderEnum, error) {
	switch strings.ToLower(raw) {
	case "google":
		return dbgen.IntegrationProviderEnumGoogle, nil
	case "microsoft":
		return dbgen.IntegrationProviderEnumMicrosoft, nil
	default:
		return "", errors.New("provider must be google or microsoft")
	}
}

func parseIntegrationType(raw string) (dbgen.IntegrationTypeEnum, error) {
	switch strings.ToLower(raw) {
	case "email":
		return dbgen.IntegrationTypeEnumEmail, nil
	case "calendar":
		return dbgen.IntegrationTypeEnumCalendar, nil
	default:
		return "", errors.New("integrationType must be email or calendar")
	}
}

func parseIntegrationStatus(raw string) (dbgen.IntegrationStatusEnum, error) {
	switch strings.ToLower(raw) {
	case "active":
		return dbgen.IntegrationStatusEnumActive, nil
	case "revoked":
		return dbgen.IntegrationStatusEnumRevoked, nil
	case "error":
		return dbgen.IntegrationStatusEnumError, nil
	default:
		return "", errors.New("status must be active, revoked, or error")
	}
}

func parseApprovalStatus(raw string) (dbgen.ApprovalStatusEnum, error) {
	switch strings.ToLower(raw) {
	case "pending":
		return dbgen.ApprovalStatusEnumPending, nil
	case "approved":
		return dbgen.ApprovalStatusEnumApproved, nil
	case "rejected":
		return dbgen.ApprovalStatusEnumRejected, nil
	default:
		return "", errors.New("status must be pending, approved, or rejected")
	}
}

func integrationConnectionDTO(row dbgen.IntegrationConnection) map[string]any {
	return map[string]any{
		"id":                pgUUIDToString(row.ID),
		"userId":            pgUUIDToString(row.UserID),
		"provider":          string(row.Provider),
		"integrationType":   string(row.IntegrationType),
		"externalAccountId": row.ExternalAccountID,
		"status":            string(row.Status),
		"scopes":            row.Scopes,
		"expiresAt":         pgTimestampToString(row.ExpiresAt),
		"updatedAt":         pgTimestampToString(row.UpdatedAt),
	}
}

func approvalDTO(row dbgen.ApprovalRequest) map[string]any {
	return map[string]any{
		"id":             pgUUIDToString(row.ID),
		"entityType":     row.EntityType,
		"entityId":       pgUUIDToString(row.EntityID),
		"requestedBy":    pgUUIDToString(row.RequestedBy),
		"approverUserId": pgUUIDToString(row.ApproverUserID),
		"status":         string(row.Status),
		"reason":         row.Reason,
		"decisionNote":   pgTextToString(row.DecisionNote),
		"decidedAt":      pgTimestampToString(row.DecidedAt),
		"createdAt":      pgTimestampToString(row.CreatedAt),
	}
}

func parseInt16(raw string, fallback int16) int16 {
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	v = max(0, min(100, v))
	return int16(v)
}

func parseFloat(raw string, fallback float64) float64 {
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return fallback
	}
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func readCSVRecords(r *http.Request) ([][]string, error) {
	contentType := r.Header.Get("Content-Type")
	var reader io.Reader = r.Body

	if strings.Contains(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return nil, err
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader = file
	}

	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true
	return csvReader.ReadAll()
}
