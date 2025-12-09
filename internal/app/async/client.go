package async

import (
	"bytes"
	"encoding/json"
	"fmt"
	"lab1/internal/app/config"
	"lab1/internal/app/ds"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type AsyncClient struct {
	config *config.Config
	logger *logrus.Logger
}

func NewAsyncClient(config *config.Config, logger *logrus.Logger) *AsyncClient {
	return &AsyncClient{
		config: config,
		logger: logger,
	}
}

// CalculationRequest запрос на расчет
type CalculationRequest struct {
	CalcID         string  `json:"calc_id"`
	StarID         string  `json:"star_id"`
	ScopeID        string  `json:"scope_id"`
	InpDist        float64 `json:"inp_dist"`
	InpTexp        float64 `json:"inp_texp"`
	InpMass        float64 `json:"inp_mass"`
	ScopeLambda    float64 `json:"scope_lambda"`
	ScopeDeltaLamb float64 `json:"scope_delta_lamb"`
	ScopeZeroPoint float64 `json:"scope_zero_point"`
}

// CalculationResult результат расчета
type CalculationResult struct {
	StarID  string  `json:"star_id"`
	ScopeID string  `json:"scope_id"`
	ResEn   float64 `json:"res_en"`
	ResNi   float64 `json:"res_ni"`
}

func (c *AsyncClient) SendCalculation(calc ds.Calc, scope ds.Scope, starID string) error {
	calcID := fmt.Sprintf("%s_%d", starID, calc.ScopeID)

	request := CalculationRequest{
		CalcID:         calcID,
		StarID:         starID,
		ScopeID:        fmt.Sprintf("%d", calc.ScopeID),
		InpDist:        calc.InpDist,
		InpTexp:        calc.InpTexp,
		InpMass:        calc.InpMass,
		ScopeLambda:    scope.Lambda,
		ScopeDeltaLamb: scope.DeltaLamb,
		ScopeZeroPoint: scope.ZeroPoint,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal calculation request: %w", err)
	}

	url := fmt.Sprintf("%s/calculate/", c.config.Async.URL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	c.logger.Infof("Sending calculation to async service: %s", url)
	c.logger.Debugf("Request data: %s", string(jsonData))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to async service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("async service returned status: %d", resp.StatusCode)
	}

	c.logger.Infof("Calculation sent successfully, calc_id: %s", calcID)
	return nil
}
