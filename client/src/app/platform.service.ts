import { Injectable } from '@angular/core';
import { HttpClient, HttpErrorResponse, HttpHeaders, HttpParams } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError, map, tap } from 'rxjs/operators';
import { environment } from '../environments/environment';
import { ApiEnvelope } from './api-envelope';
import { EconomyLedgerListData } from './economy/economy.models';
import { MessageService } from './message.service';

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

  getBackupsStatus(): Observable<unknown> {
    return this.http.get<ApiEnvelope<unknown>>(`${this.base}/backups/status`).pipe(
      map((e) => e.data),
      tap(() => this.log('Fetched backups status')),
      catchError(this.handleError<unknown>('getBackupsStatus', null))
    );
  }

  /** GET /economy/ledger — Phase B read-only; requires economy.read */
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
            this.messageService.add(
              'PlatformService: economy ledger requires economy.read (operator/support/security_admin)'
            );
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

  getAdminAuditEvents(): Observable<unknown> {
    return this.http.get<ApiEnvelope<unknown>>(`${this.base}/audit/admin-events`).pipe(
      map((e) => e.data),
      tap(() => this.log('Fetched admin audit events')),
      catchError(this.handleError<unknown>('getAdminAuditEvents', null))
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
    this.messageService.add(`PlatformService: ${message}`);
  }
}
