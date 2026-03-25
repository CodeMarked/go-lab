import { Injectable } from '@angular/core';
import { HttpClient, HttpErrorResponse, HttpHeaders, HttpParams } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError, map, tap } from 'rxjs/operators';
import { environment } from '../environments/environment';
import { ApiEnvelope } from './api-envelope';
import { EconomyLedgerListData } from './economy/economy.models';
import {
  BackupRestoreCreateData,
  BackupRestoreIdStatus,
  BackupRestoreListData,
  BackupsStatusData,
  SecurityMeData
} from './dataops/dataops.models';
import { MessageService } from './message.service';
import {
  OperatorCaseActionsData,
  OperatorCaseListData,
  OperatorCaseNotesData,
  OperatorCaseRow
} from './cases/cases.models';

/** Must match api/middleware/privileged.go */
export const PLATFORM_ACTION_REASON_HEADER = 'X-Platform-Action-Reason';

@Injectable({
  providedIn: 'root'
})
export class PlatformService {
  private readonly base = `${environment.apiBaseUrl}/api/v1`;

  constructor(
    private http: HttpClient,
    private messageService: MessageService
  ) {}

  getPlayers(): Observable<unknown> {
    return this.http.get<ApiEnvelope<unknown>>(`${this.base}/players`).pipe(
      map((e) => e.data),
      tap(() => this.log('Fetched players stub')),
      catchError(this.handleError<unknown>('getPlayers', null))
    );
  }

  getCharacters(): Observable<unknown> {
    return this.http.get<ApiEnvelope<unknown>>(`${this.base}/characters`).pipe(
      map((e) => e.data),
      tap(() => this.log('Fetched characters stub')),
      catchError(this.handleError<unknown>('getCharacters', null))
    );
  }

  /** GET /backups/status — requires backups.read */
  getBackupsStatus(): Observable<BackupsStatusData | null> {
    return this.http.get<ApiEnvelope<BackupsStatusData>>(`${this.base}/backups/status`).pipe(
      map((e) => e.data),
      tap(() => this.log('Fetched backups status')),
      catchError(this.handleError<BackupsStatusData | null>('getBackupsStatus', null))
    );
  }

  /** GET /backups/restore-requests — requires backups.read */
  listBackupRestoreRequests(filters: {
    status?: string;
    limit?: number;
  }): Observable<BackupRestoreListData | null> {
    let params = new HttpParams();
    if (filters.status != null && filters.status.trim() !== '') {
      params = params.set('status', filters.status.trim());
    }
    if (filters.limit != null && !Number.isNaN(filters.limit)) {
      params = params.set('limit', String(filters.limit));
    }
    return this.http
      .get<ApiEnvelope<BackupRestoreListData>>(`${this.base}/backups/restore-requests`, { params })
      .pipe(
        map((e) => e.data),
        tap(() => this.log('Listed backup restore requests')),
        catchError(this.handleError<BackupRestoreListData | null>('listBackupRestoreRequests', null))
      );
  }

  /** POST /backups/restore-requests — requires backups.restore.request + reason header */
  postBackupRestoreRequest(
    body: { scope: string; restore_point_label: string },
    actionReason: string
  ): Observable<BackupRestoreCreateData | null> {
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      [PLATFORM_ACTION_REASON_HEADER]: actionReason
    });
    return this.http
      .post<ApiEnvelope<BackupRestoreCreateData>>(`${this.base}/backups/restore-requests`, body, {
        headers
      })
      .pipe(
        map((e) => e.data),
        tap(() => this.log('Created backup restore request')),
        catchError(this.handleError<BackupRestoreCreateData | null>('postBackupRestoreRequest', null))
      );
  }

  postBackupRestoreApprove(id: number, actionReason: string): Observable<BackupRestoreIdStatus | null> {
    return this.postBackupRestoreAction(id, 'approve', actionReason);
  }

  postBackupRestoreReject(id: number, actionReason: string): Observable<BackupRestoreIdStatus | null> {
    return this.postBackupRestoreAction(id, 'reject', actionReason);
  }

  postBackupRestoreFulfill(id: number, actionReason: string): Observable<BackupRestoreIdStatus | null> {
    return this.postBackupRestoreAction(id, 'fulfill', actionReason);
  }

  postBackupRestoreCancel(id: number, actionReason: string): Observable<BackupRestoreIdStatus | null> {
    return this.postBackupRestoreAction(id, 'cancel', actionReason);
  }

  private postBackupRestoreAction(
    id: number,
    suffix: 'approve' | 'reject' | 'fulfill' | 'cancel',
    actionReason: string
  ): Observable<BackupRestoreIdStatus | null> {
    const headers = new HttpHeaders({
      [PLATFORM_ACTION_REASON_HEADER]: actionReason
    });
    const url = `${this.base}/backups/restore-requests/${id}/${suffix}`;
    return this.http.post<ApiEnvelope<BackupRestoreIdStatus>>(url, {}, { headers }).pipe(
      map((e) => e.data),
      tap(() => this.log(`Backup restore ${suffix} ok`)),
      catchError(this.handleError<BackupRestoreIdStatus | null>(`postBackupRestore${suffix}`, null))
    );
  }

  /** GET /economy/ledger — read-only; requires economy.read */
  getEconomyLedger(filters: {
    limit?: number;
    platform_user_id?: number;
    event_type?: string;
    from?: string;
    to?: string;
    before_id?: number;
  }): Observable<EconomyLedgerListData | null> {
    let params = new HttpParams();
    if (filters.limit != null && !Number.isNaN(filters.limit)) {
      params = params.set('limit', String(filters.limit));
    }
    if (filters.platform_user_id != null && filters.platform_user_id > 0) {
      params = params.set('platform_user_id', String(filters.platform_user_id));
    }
    if (filters.event_type != null && filters.event_type.trim() !== '') {
      params = params.set('event_type', filters.event_type.trim());
    }
    if (filters.from != null && filters.from.trim() !== '') {
      params = params.set('from', filters.from.trim());
    }
    if (filters.to != null && filters.to.trim() !== '') {
      params = params.set('to', filters.to.trim());
    }
    if (filters.before_id != null && filters.before_id > 0) {
      params = params.set('before_id', String(filters.before_id));
    }
    return this.http
      .get<ApiEnvelope<EconomyLedgerListData>>(`${this.base}/economy/ledger`, { params })
      .pipe(
        map((e) => e.data),
        tap(() => this.log('Fetched economy ledger')),
        catchError((error: unknown) => {
          if (error instanceof HttpErrorResponse && error.status === 403) {
            this.messageService.add("Couldn't load economy ledger (permission required).");
            return of(null);
          }
          return this.handleError<EconomyLedgerListData | null>('getEconomyLedger', null)(error);
        })
      );
  }

  getSecurityMe(): Observable<unknown> {
    return this.http.get<ApiEnvelope<unknown>>(`${this.base}/security/me`).pipe(
      map((e) => e.data),
      tap(() => this.log('Fetched security/me')),
      catchError(this.handleError<unknown>('getSecurityMe', null))
    );
  }

  /** Typed roles + effective_permissions for UI gating */
  getSecurityMeTyped(): Observable<SecurityMeData | null> {
    return this.http.get<ApiEnvelope<SecurityMeData>>(`${this.base}/security/me`).pipe(
      map((e) => e.data),
      tap(() => this.log('Fetched security/me')),
      catchError(this.handleError<SecurityMeData | null>('getSecurityMeTyped', null))
    );
  }

  getAdminAuditEvents(): Observable<unknown> {
    return this.http.get<ApiEnvelope<unknown>>(`${this.base}/audit/admin-events`).pipe(
      map((e) => e.data),
      tap(() => this.log('Fetched admin audit events')),
      catchError(this.handleError<unknown>('getAdminAuditEvents', null))
    );
  }

  listOperatorCases(filters: {
    limit?: number;
    status?: string;
    subject_platform_user_id?: number;
    before_id?: number;
  }): Observable<OperatorCaseListData | null> {
    let params = new HttpParams();
    if (filters.limit != null && !Number.isNaN(filters.limit)) {
      params = params.set('limit', String(filters.limit));
    }
    if (filters.status != null && filters.status.trim() !== '') {
      params = params.set('status', filters.status.trim());
    }
    if (filters.subject_platform_user_id != null && filters.subject_platform_user_id > 0) {
      params = params.set('subject_platform_user_id', String(filters.subject_platform_user_id));
    }
    if (filters.before_id != null && filters.before_id > 0) {
      params = params.set('before_id', String(filters.before_id));
    }
    return this.http
      .get<ApiEnvelope<OperatorCaseListData>>(`${this.base}/cases`, { params })
      .pipe(
        map((e) => e.data),
        catchError(this.handleError<OperatorCaseListData | null>('Load cases', null))
      );
  }

  getOperatorCase(id: number): Observable<OperatorCaseRow | null> {
    return this.http.get<ApiEnvelope<OperatorCaseRow>>(`${this.base}/cases/${id}`).pipe(
      map((e) => e.data),
      catchError(this.handleError<OperatorCaseRow | null>('Load case', null))
    );
  }

  postOperatorCase(
    body: {
      subject_platform_user_id: number;
      title: string;
      description?: string;
      priority?: string;
      subject_character_ref?: string;
    },
    actionReason: string
  ): Observable<{ id: number; status: string } | null> {
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      [PLATFORM_ACTION_REASON_HEADER]: actionReason
    });
    return this.http
      .post<ApiEnvelope<{ id: number; status: string }>>(`${this.base}/cases`, body, { headers })
      .pipe(
        map((e) => e.data),
        catchError(this.handleError<{ id: number; status: string } | null>('Create case', null))
      );
  }

  patchOperatorCase(
    id: number,
    body: { status?: string; priority?: string; assigned_to_user_id?: number | null },
    actionReason: string
  ): Observable<OperatorCaseRow | null> {
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      [PLATFORM_ACTION_REASON_HEADER]: actionReason
    });
    return this.http
      .patch<ApiEnvelope<OperatorCaseRow>>(`${this.base}/cases/${id}`, body, { headers })
      .pipe(
        map((e) => e.data),
        catchError(this.handleError<OperatorCaseRow | null>('Update case', null))
      );
  }

  listOperatorCaseNotes(caseId: number): Observable<OperatorCaseNotesData | null> {
    return this.http
      .get<ApiEnvelope<OperatorCaseNotesData>>(`${this.base}/cases/${caseId}/notes`)
      .pipe(
        map((e) => e.data),
        catchError(this.handleError<OperatorCaseNotesData | null>('Load notes', null))
      );
  }

  postOperatorCaseNote(
    caseId: number,
    body: { body: string },
    actionReason: string
  ): Observable<{ id: number; case_id: number } | null> {
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      [PLATFORM_ACTION_REASON_HEADER]: actionReason
    });
    return this.http
      .post<ApiEnvelope<{ id: number; case_id: number }>>(`${this.base}/cases/${caseId}/notes`, body, {
        headers
      })
      .pipe(
        map((e) => e.data),
        catchError(this.handleError<{ id: number; case_id: number } | null>('Add note', null))
      );
  }

  listOperatorCaseActions(caseId: number): Observable<OperatorCaseActionsData | null> {
    return this.http
      .get<ApiEnvelope<OperatorCaseActionsData>>(`${this.base}/cases/${caseId}/actions`)
      .pipe(
        map((e) => e.data),
        catchError(this.handleError<OperatorCaseActionsData | null>('Load actions', null))
      );
  }

  postOperatorCaseSanction(
    caseId: number,
    body: { sanction_type: string; expires_at?: string },
    actionReason: string
  ): Observable<{ ok: boolean; recorded: string } | null> {
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      [PLATFORM_ACTION_REASON_HEADER]: actionReason
    });
    return this.http
      .post<ApiEnvelope<{ ok: boolean; recorded: string }>>(
        `${this.base}/cases/${caseId}/sanctions`,
        body,
        { headers }
      )
      .pipe(
        map((e) => e.data),
        catchError(this.handleError<{ ok: boolean; recorded: string } | null>('Record sanction', null))
      );
  }

  postOperatorCaseRecoveryRequest(
    caseId: number,
    body: { character_ref?: string },
    actionReason: string
  ): Observable<{ ok: boolean; recorded: string } | null> {
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      [PLATFORM_ACTION_REASON_HEADER]: actionReason
    });
    return this.http
      .post<ApiEnvelope<{ ok: boolean; recorded: string }>>(
        `${this.base}/cases/${caseId}/recovery-requests`,
        body,
        { headers }
      )
      .pipe(
        map((e) => e.data),
        catchError(
          this.handleError<{ ok: boolean; recorded: string } | null>('Recovery request', null)
        )
      );
  }

  postOperatorCaseAppealResolve(
    caseId: number,
    body: { outcome: string; notes?: string },
    actionReason: string
  ): Observable<{ ok: boolean; recorded: string } | null> {
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      [PLATFORM_ACTION_REASON_HEADER]: actionReason
    });
    return this.http
      .post<ApiEnvelope<{ ok: boolean; recorded: string }>>(
        `${this.base}/cases/${caseId}/appeals/resolve`,
        body,
        { headers }
      )
      .pipe(
        map((e) => e.data),
        catchError(
          this.handleError<{ ok: boolean; recorded: string } | null>('Resolve appeal', null)
        )
      );
  }

  postSupportAck(reason: string, message?: string): Observable<unknown> {
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
      [PLATFORM_ACTION_REASON_HEADER]: reason
    });
    const body = message?.trim() ? { message: message.trim() } : {};
    return this.http
      .post<ApiEnvelope<unknown>>(`${this.base}/support/ack`, body, { headers })
      .pipe(
        map((e) => e.data),
        tap(() => this.log('Posted support ack')),
        catchError(this.handleError<unknown>('postSupportAck', null))
      );
  }

  private handleError<T>(operation: string, result: T) {
    return (error: unknown): Observable<T> => {
      console.error(error);
      let msg = error instanceof Error ? error.message : String(error);
      if (error instanceof HttpErrorResponse) {
        const apiErr = error.error?.error;
        const apiMsg =
          apiErr && typeof apiErr.message === 'string' ? apiErr.message : '';
        const apiCode = apiErr && typeof apiErr.code === 'string' ? apiErr.code : '';
        msg = `${error.status} ${error.statusText}`.trim();
        if (apiCode || apiMsg) {
          msg = `${msg} ${apiCode} ${apiMsg}`.trim();
        }
      }
      this.log(`${operation} failed: ${msg}`);
      return of(result as T);
    };
  }

  private log(message: string) {
    this.messageService.add(message);
  }
}
