export type AssetPricePoint = {
  date: string;
  open: number;
  high: number;
  low: number;
  close: number;
  adjustedClose: number;
  volume: number;
  /** Simple daily return from adjusted close; null/omitted when undefined in DB. */
  dailyReturn?: number | null;
  /** Five trading-day rolling simple return; null/omitted when undefined in DB. */
  rolling5DayReturn?: number | null;
};

export type AssetPricesPayload = {
  symbol: string;
  points: AssetPricePoint[];
};
