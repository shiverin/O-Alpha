interface LivePosition {
  symbol: string;
  qty: number;
  avg_entry_price: number;
  current_price: number;
  unrealized_pnl: number;
  exposure: number;
}

interface LiveSummary {
  total_asset_value: number;
  change_percent_24h: number;
  change_dollar_24h: number;
  estimated_annual_yield: number;
  target_progress_percent: number;
  timestamp: string;
}

interface LiveHistoryPoint {
  total_asset_value: number;
  timestamp: string;
  change_dollar_24h?: number;
}

export interface LivePricePatch {
  symbol: string;
  price: number;
  timestamp: string;
}

export function applyLivePriceToPositions<T extends LivePosition>(
  positions: T[] | undefined,
  patch: LivePricePatch,
) {
  if (!positions || patch.price <= 0) {
    return { positions, deltaExposure: 0 };
  }

  let deltaExposure = 0;
  const symbol = patch.symbol.toUpperCase();
  const nextPositions = positions.map((position) => {
    if (position.symbol.toUpperCase() !== symbol) {
      return position;
    }

    const oldExposure =
      position.exposure || position.qty * position.current_price;
    const exposure = position.qty * patch.price;
    deltaExposure += exposure - oldExposure;

    return {
      ...position,
      current_price: patch.price,
      exposure,
      unrealized_pnl: (patch.price - position.avg_entry_price) * position.qty,
    };
  });

  return { positions: nextPositions, deltaExposure };
}

export function applyLivePriceToSummary<T extends LiveSummary>(
  summary: T | undefined,
  deltaExposure: number,
  timestamp: string,
) {
  if (!summary || deltaExposure === 0) {
    return summary;
  }

  const totalAssetValue = summary.total_asset_value + deltaExposure;
  const changeDollar24h = summary.change_dollar_24h + deltaExposure;
  const baseline = summary.total_asset_value - summary.change_dollar_24h;
  const changePercent24h =
    baseline > 0 ? (changeDollar24h / baseline) * 100 : 0;
  const targetValue =
    summary.target_progress_percent > 0
      ? summary.total_asset_value / (summary.target_progress_percent / 100)
      : 0;
  const targetProgressPercent =
    targetValue > 0 ? (totalAssetValue / targetValue) * 100 : 0;

  return {
    ...summary,
    total_asset_value: totalAssetValue,
    change_dollar_24h: changeDollar24h,
    change_percent_24h: clampPercent(changePercent24h),
    target_progress_percent: clampPercent(targetProgressPercent),
    timestamp,
  };
}

export function applyLivePriceToHistory<T extends LiveHistoryPoint>(
  history: T[] | undefined,
  deltaExposure: number,
  timestamp: string,
) {
  if (!history || history.length === 0 || deltaExposure === 0) {
    return history;
  }

  const nextHistory = [...history];
  const latest = nextHistory[nextHistory.length - 1];
  nextHistory[nextHistory.length - 1] = {
    ...latest,
    total_asset_value: latest.total_asset_value + deltaExposure,
    ...(latest.change_dollar_24h === undefined
      ? {}
      : { change_dollar_24h: latest.change_dollar_24h + deltaExposure }),
    timestamp,
  };
  return nextHistory.slice(-30);
}

function clampPercent(value: number) {
  if (value > 999.99) return 999.99;
  if (value < -999.99) return -999.99;
  return value;
}
