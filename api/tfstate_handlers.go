package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hypertf/nahcloud-server/domain"
)

// TFStateGet handles GET /v1/tfstate/{state_id}
func (h *Handler) TFStateGet(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]

	state, err := h.service.GetTFState(id)
	if err != nil {
		if domain.IsNotFound(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		h.writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(state))
}

// TFStatePost handles POST /v1/tfstate/{state_id}
func (h *Handler) TFStatePost(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]

	// Enforce lock if present
	if rawLock, lockInfo, err := h.service.GetTFStateLock(id); err == nil && lockInfo != nil {
		provided := r.URL.Query().Get("ID")
		if provided == "" || provided != lockInfo.ID {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusLocked) // 423
			w.Write([]byte(rawLock))
			return
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, domain.InternalError("failed to read request body"))
		return
	}
	if err := h.service.SetTFState(id, string(body)); err != nil {
		h.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// TFStateDelete handles DELETE /v1/tfstate/{state_id}
func (h *Handler) TFStateDelete(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]

	// Enforce lock if present
	if rawLock, lockInfo, err := h.service.GetTFStateLock(id); err == nil && lockInfo != nil {
		provided := r.URL.Query().Get("ID")
		if provided == "" || provided != lockInfo.ID {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusLocked) // 423
			w.Write([]byte(rawLock))
			return
		}
	}

	if err := h.service.DeleteTFState(id); err != nil {
		h.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// TFStateLock handles LOCK /v1/tfstate/{state_id}
func (h *Handler) TFStateLock(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, domain.InternalError("failed to read request body"))
		return
	}
	// Validate lock JSON minimally to extract ID
	var li domain.TFStateLock
	if err := json.Unmarshal(body, &li); err != nil || li.ID == "" {
		h.writeError(w, domain.InvalidInputError("invalid lock payload: missing or invalid ID", nil))
		return
	}

	// Try to place the lock
	locked, existing, err := h.service.TryLockTFState(id, string(body))
	if err != nil {
		h.writeError(w, err)
		return
	}
	if locked {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusLocked) // 423
		w.Write([]byte(existing))
		return
	}
	w.WriteHeader(http.StatusOK)
}

// TFStateUnlock handles UNLOCK /v1/tfstate/{state_id}
func (h *Handler) TFStateUnlock(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		h.writeError(w, err)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]

	provided := r.URL.Query().Get("ID")
	// Get current lock
	rawLock, lockInfo, err := h.service.GetTFStateLock(id)
	if err != nil {
		// Not locked
		w.WriteHeader(http.StatusOK)
		return
	}
	if lockInfo == nil || provided == lockInfo.ID {
		// No parsed info or matching ID: unlock
		if _, _, err := h.service.UnlockTFState(id); err != nil {
			h.writeError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	// Mismatch
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusConflict) // 409
	w.Write([]byte(rawLock))
}
