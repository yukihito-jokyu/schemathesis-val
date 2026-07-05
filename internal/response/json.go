package response

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/yukihito-jokyu/schemathesis-val/internal/model"
)

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func Error(w http.ResponseWriter, status int, code model.ErrorCode, message string, details []string) {
	resp := model.ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
	JSON(w, status, resp)
}

func DecodeAndValidate(w http.ResponseWriter, r *http.Request, dst interface{}, allow400 bool) ([]byte, bool) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		status := http.StatusUnprocessableEntity
		if allow400 {
			status = http.StatusBadRequest
		}
		Error(w, status, model.CodeBadRequest, "failed to read body", []string{err.Error()})
		return nil, false
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &raw); err == nil {
		var nullFields []string
		for k, v := range raw {
			if v == nil {
				nullFields = append(nullFields, fmt.Sprintf("%s cannot be null", k))
			}
		}
		if len(nullFields) > 0 {
			Error(w, http.StatusUnprocessableEntity, model.CodeValidationError, "validation failed", nullFields)
			return nil, false
		}
	}

	decoder := json.NewDecoder(bytes.NewReader(bodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		status := http.StatusUnprocessableEntity
		if allow400 {
			status = http.StatusBadRequest
		}
		Error(w, status, model.CodeBadRequest, "invalid JSON or unknown properties", []string{err.Error()})
		return nil, false
	}

	return bodyBytes, true
}
