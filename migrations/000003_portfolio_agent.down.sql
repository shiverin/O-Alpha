ALTER TABLE agent_runs DROP CONSTRAINT IF EXISTS chk_agent_runs_strategy_type;
ALTER TABLE agent_runs ADD CONSTRAINT chk_agent_runs_strategy_type
CHECK (strategy_type IN (
    'MA_CROSSOVER', 'KALMAN', 'REGIME_DETECTOR', 'PAIRS',
    'HMM_ENSEMBLE', 'ML_META_LABEL', 'XSEC_MOMENTUM',
    'KALMAN_COINTEGRATION', 'MULTI_ENGINE_ENSEMBLE'
));
