package agent

import "github.com/oalpha/internal/agent/portfolio"

type ExecutionRouter = portfolio.ExecutionRouter
type PortfolioPaperAccount = portfolio.PortfolioPaperAccount
type PortfolioPaperPosition = portfolio.PortfolioPaperPosition
type PortfolioAgentWorker = portfolio.PortfolioAgentWorker
type PortfolioAgentManager = portfolio.PortfolioAgentManager
type StrategyRiskProfile = portfolio.StrategyRiskProfile
type StrategyDeploymentStatus = portfolio.StrategyDeploymentStatus
type StrategyCatalogConfig = portfolio.StrategyCatalogConfig
type StrategySpec = portfolio.StrategySpec

const (
	StrategyRiskLow    = portfolio.StrategyRiskLow
	StrategyRiskMedium = portfolio.StrategyRiskMedium
	StrategyRiskHigh   = portfolio.StrategyRiskHigh

	StrategyStatusPromotedResearch    = portfolio.StrategyStatusPromotedResearch
	StrategyStatusConservativeVariant = portfolio.StrategyStatusConservativeVariant
	StrategyStatusExperimentalVariant = portfolio.StrategyStatusExperimentalVariant
	StrategyStatusRejectedDiagnostic  = portfolio.StrategyStatusRejectedDiagnostic
	StrategyStatusPaperOnly           = portfolio.StrategyStatusPaperOnly
)

var NewPortfolioPaperAccount = portfolio.NewPortfolioPaperAccount
var NewPortfolioAgentWorker = portfolio.NewPortfolioAgentWorker
var NewPortfolioAgentManager = portfolio.NewPortfolioAgentManager
var DefaultStrategyCatalogConfig = portfolio.DefaultStrategyCatalogConfig
var AvailableStrategySpecs = portfolio.AvailableStrategySpecs
var StrategySpecByKey = portfolio.StrategySpecByKey
var NewStrategyFromCatalog = portfolio.NewStrategyFromCatalog
