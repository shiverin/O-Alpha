package backtest

// KalmanFilter1D represents a one-dimensional Kalman Filter state.
type KalmanFilter1D struct {
	Estimate         float64 // x: The current estimated true price
	ErrorCov         float64 // P: The estimation error covariance
	ProcessNoise     float64 // Q: Environment/process variance (system dynamics)
	MeasurementNoise float64 // R: Measurement variance (market micro-noise)
	isInitialized    bool
}

// NewKalmanFilter1D initializes a new 1D Kalman Filter.
func NewKalmanFilter1D(processNoise, measurementNoise float64) *KalmanFilter1D {
	return &KalmanFilter1D{
		ProcessNoise:     processNoise,
		MeasurementNoise: measurementNoise,
		ErrorCov:         1.0, // Start with a baseline error covariance
	}
}

// Update runs the Prediction and Correction cycles given a new price measurement.
func (kf *KalmanFilter1D) Update(measurement float64) float64 {
	if !kf.isInitialized {
		kf.Estimate = measurement
		kf.isInitialized = true
		return kf.Estimate
	}

	// 1. PREDICT PHASE
	// x_pred = x
	// P_pred = P + Q
	predEstimate := kf.Estimate
	predErrorCov := kf.ErrorCov + kf.ProcessNoise

	// 2. UPDATE / CORRECT PHASE
	// Calculate Kalman Gain: K = P_pred / (P_pred + R)
	kalmanGain := predErrorCov / (predErrorCov + kf.MeasurementNoise)

	// Update Estimate: x = x_pred + K * (measurement - x_pred)
	kf.Estimate = predEstimate + kalmanGain*(measurement-predEstimate)

	// Update Error Covariance: P = (1 - K) * P_pred
	kf.ErrorCov = (1.0 - kalmanGain) * predErrorCov

	return kf.Estimate
}
