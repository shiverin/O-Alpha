export const mockExecutionLogs = [
  {
    time: "14:32:01",
    asset: "BTC-USD",
    side: "BUY",
    price: "$64,210.50",
    primary: true,
  },
  {
    time: "14:31:45",
    asset: "ETH-USD",
    side: "SELL",
    price: "$3,450.20",
    primary: false,
  },
  {
    time: "14:30:12",
    asset: "AAPL",
    side: "BUY",
    price: "$175.40",
    highlight: true,
  },
  {
    time: "14:28:55",
    asset: "TSLA",
    side: "SELL",
    price: "$210.05",
    primary: false,
  },
  {
    time: "14:25:33",
    asset: "SOL-USD",
    side: "BUY",
    price: "$145.60",
    primary: true,
  },
  {
    time: "14:20:10",
    asset: "MSFT",
    side: "BUY",
    price: "$410.25",
    primary: false,
  },
];

export const allocationSegments = [
  {
    label: "Equities",
    percentage: 40,
    value: "$960K",
    color: "#00f0ff",
    glowClass: "bg-primary-container shadow-[0_0_12px_rgba(0,240,255,0.4)]",
  },
  {
    label: "Crypto",
    percentage: 30,
    value: "$720K",
    color: "#ffd700",
    glowClass: "bg-secondary-fixed shadow-[0_0_12px_rgba(255,215,0,0.3)]",
  },
  {
    label: "Fixed Inc.",
    percentage: 20,
    value: "$480K",
    color: "#353535",
    glowClass: "bg-surface-container-highest",
  },
  {
    label: "Cash",
    percentage: 10,
    value: "$240K",
    color: "#1c1b1b",
    glowClass: "bg-surface-container-low border border-outline-variant/30",
  },
];

export interface PortfolioSummary {
  totalAssetValue: number;
  changePercent24h: number;
  changeDollar24h: number;
  sparklinePath: string;
}

export interface CompositionSegment {
  label: string;
  percentage: number;
  color: string;
  glowClass: string;
  dashOffset: number;
  rotation: number;
}

export interface PortfolioMetrics {
  topPerformer: {
    symbol: string;
    contribution: number;
  };
  riskProfile: {
    label: string;
    sharpeRatio: number;
  };
  estimatedAnnualYield: number;
  targetProgressPercent: number;
}

export interface AssetPosition {
  symbol: string;
  name: string;
  category: "Equity" | "Crypto" | "Fixed Inc" | "Cash";
  initials: string;
  allocation: number;
  currentPrice: number;
  unrealizedPnL: number;
  exposure: number;
  isPositive: boolean;
  borderClass?: string;
}

export const portfolioSummary: PortfolioSummary = {
  totalAssetValue: 2481903.5,
  changePercent24h: 4.2,
  changeDollar24h: 104239.95,
  sparklinePath: "M0 80 Q 50 70, 100 85 T 200 60 T 300 40 T 400 20",
};

export const compositionSegments: CompositionSegment[] = [
  {
    label: "Equities",
    percentage: 60,
    color: "#00dbe9",
    glowClass: "bg-primary-fixed-dim shadow-[0_0_8px_rgba(0,219,233,0.5)]",
    dashOffset: 251,
    rotation: 0,
  },
  {
    label: "Crypto Assets",
    percentage: 30,
    color: "#e9c400",
    glowClass: "bg-secondary-fixed-dim shadow-[0_0_8px_rgba(233,196,0,0.3)]",
    dashOffset: 439,
    rotation: 216,
  },
  {
    label: "Cash & Equiv",
    percentage: 10,
    color: "#849495",
    glowClass: "bg-outline",
    dashOffset: 565,
    rotation: 324,
  },
];

export const portfolioMetrics: PortfolioMetrics = {
  topPerformer: {
    symbol: "NVDA",
    contribution: 12.4,
  },
  riskProfile: {
    label: "Moderate",
    sharpeRatio: 1.84,
  },
  estimatedAnnualYield: 124500.0,
  targetProgressPercent: 85,
};

export const assetPositions: AssetPosition[] = [
  {
    symbol: "NVDA",
    name: "NVIDIA Corp",
    category: "Equity",
    initials: "NV",
    allocation: 18.5,
    currentPrice: 134.2,
    unrealizedPnL: 45200.0,
    exposure: 459152.0,
    isPositive: true,
  },
  {
    symbol: "BTC",
    name: "Bitcoin",
    category: "Crypto",
    initials: "BT",
    allocation: 15.0,
    currentPrice: 64230.0,
    unrealizedPnL: 12450.0,
    exposure: 372285.0,
    isPositive: true,
  },
  {
    symbol: "AAPL",
    name: "Apple Inc",
    category: "Equity",
    initials: "AP",
    allocation: 12.2,
    currentPrice: 189.5,
    unrealizedPnL: -2100.5,
    exposure: 302789.0,
    isPositive: false,
  },
  {
    symbol: "ETH",
    name: "Ethereum",
    category: "Crypto",
    initials: "ET",
    allocation: 8.4,
    currentPrice: 3450.2,
    unrealizedPnL: 8900.0,
    exposure: 208483.0,
    isPositive: true,
    borderClass: "border-secondary-fixed-dim/30 text-secondary-fixed-dim",
  },
];

const fallbackData = {
  mockExecutionLogs,
  allocationSegments,
  portfolioSummary,
  compositionSegments,
  portfolioMetrics,
  assetPositions,
};

export default fallbackData;

export interface ExecutionStreamLog {
  timestamp: string;
  action:
    | "BUY_MKT"
    | "SELL_LMT"
    | "REBALANCE"
    | "BUY_LMT"
    | "SELL_MKT"
    | "CANCEL";
  asset: string;
  price: string;
  size: string;
  slippage: string;
  status: "FILLED" | "PENDING" | "COMPLETE" | "CANCELLED";
  statusColorClass: string;
  actionColorClass: string;
}

export interface SystemAlertItem {
  id: string;
  title: string;
  description: string;
  timeLabel: string;
  iconName: "trending_down" | "api";
  borderClass: string;
}

export interface AgentLogicNode {
  id: string;
  timeLabel: string;
  title: string;
  description: string;
  isCurrent: boolean;
}

export const mockExecutionStream: ExecutionStreamLog[] = [
  {
    timestamp: "14:22:05.102",
    action: "BUY_MKT",
    asset: "ETH/USD",
    price: "$3,421.50",
    size: "45.00",
    slippage: "0.01%",
    status: "FILLED",
    statusColorClass:
      "bg-primary-fixed-dim/10 text-primary-fixed-dim border-primary-fixed-dim/20",
    actionColorClass: "text-primary-fixed-dim font-bold",
  },
  {
    timestamp: "14:21:45.881",
    action: "SELL_LMT",
    asset: "BTC/USD",
    price: "$64,102.00",
    size: "2.50",
    slippage: "-",
    status: "PENDING",
    statusColorClass:
      "bg-secondary-container/10 text-secondary-fixed border-secondary-container/20",
    actionColorClass: "text-error font-bold",
  },
  {
    timestamp: "14:15:00.000",
    action: "REBALANCE",
    asset: "PORTFOLIO",
    price: "-",
    size: "-",
    slippage: "-",
    status: "COMPLETE",
    statusColorClass:
      "bg-white/[0.04] text-on-surface/70 border-outline-variant/30",
    actionColorClass: "text-secondary-fixed font-medium",
  },
  {
    timestamp: "14:02:11.455",
    action: "BUY_LMT",
    asset: "SOL/USD",
    price: "$145.20",
    size: "1000.00",
    slippage: "0.00%",
    status: "FILLED",
    statusColorClass:
      "bg-primary-fixed-dim/10 text-primary-fixed-dim border-primary-fixed-dim/20",
    actionColorClass: "text-primary-fixed-dim font-bold",
  },
  {
    timestamp: "13:58:44.901",
    action: "SELL_MKT",
    asset: "AVAX/USD",
    price: "$35.80",
    size: "500.00",
    slippage: "-0.15%",
    status: "FILLED",
    statusColorClass:
      "bg-primary-fixed-dim/10 text-primary-fixed-dim border-primary-fixed-dim/20",
    actionColorClass: "text-error font-bold",
  },
  {
    timestamp: "13:45:12.003",
    action: "CANCEL",
    asset: "BTC/USD",
    price: "$64,500.00",
    size: "1.00",
    slippage: "-",
    status: "CANCELLED",
    statusColorClass: "bg-error/10 text-error border-error/20",
    actionColorClass: "text-error/70 font-medium",
  },
];

export const mockSystemAlerts: SystemAlertItem[] = [
  {
    id: "alert-1",
    title: "VOLATILITY SPIKE DETECTED",
    description:
      "BTC/USD volatility exceeded 3-sigma threshold. Agent shifted to risk-off posture automatically.",
    timeLabel: "12 mins ago",
    iconName: "trending_down",
    borderClass: "border-error/20 border-l-error",
  },
  {
    id: "alert-2",
    title: "API RATE LIMIT WARNING",
    description:
      "Exchange connection approaching rate limits. Execution pacing adjusted.",
    timeLabel: "45 mins ago",
    iconName: "api",
    borderClass: "border-secondary-container/20 border-l-secondary-fixed",
  },
];

export const mockAgentLogicStates: AgentLogicNode[] = [
  {
    id: "state-1",
    timeLabel: "CURRENT POSTURE",
    title: "Defensive Consolidation",
    description:
      "Reduced leverage across high-beta assets. Increasing stablecoin reserves to 40%.",
    isCurrent: true,
  },
  {
    id: "state-2",
    timeLabel: "10:00 AM UTC",
    title: "Alpha Signal Confirmed",
    description:
      "Long position initiated on ETH following network volume surge.",
    isCurrent: false,
  },
  {
    id: "state-3",
    timeLabel: "08:15 AM UTC",
    title: "Routine Portfolio Rebalance",
    description: "Weights adjusted to target allocation. Drift corrected.",
    isCurrent: false,
  },
];
