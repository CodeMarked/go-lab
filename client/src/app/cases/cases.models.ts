export interface OperatorCaseRow {
  id: number;
  created_at: string;
  updated_at: string;
  status: string;
  priority: string;
  subject_platform_user_id: number;
  subject_character_ref?: string;
  title: string;
  description?: string;
  created_by_user_id: number;
  assigned_to_user_id?: number;
}

export interface OperatorCaseListData {
  items: OperatorCaseRow[];
  limit: number;
}

export interface OperatorCaseNoteRow {
  id: number;
  case_id: number;
  created_at: string;
  body: string;
  created_by_user_id: number;
}

export interface OperatorCaseNotesData {
  items: OperatorCaseNoteRow[];
}

export interface OperatorCaseActionRow {
  id: number;
  case_id: number;
  created_at: string;
  action_kind: string;
  payload?: Record<string, unknown>;
  reason?: string;
  actor_user_id: number;
}

export interface OperatorCaseActionsData {
  items: OperatorCaseActionRow[];
}
