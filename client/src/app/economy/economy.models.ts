/** Response shape for GET /api/v1/economy/ledger */
export interface EconomyLedgerEvent {
  id: number;
  created_at: string;
  platform_user_id: number;
  event_type: string;
  amount_delta: number;
  currency_code: string;
  reference_type?: string;
  reference_id?: string;
  meta?: Record<string, unknown>;
}

export interface EconomyLedgerListData {
  items: EconomyLedgerEvent[];
  limit: number;
}
