import { Injectable } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { BehaviorSubject, Observable, Subscription, interval, of } from 'rxjs';
import { catchError, exhaustMap, map, tap } from 'rxjs/operators';
import { environment } from '../environments/environment';
import { ApiEnvelope, TokenResponseData } from './api-envelope';

export interface LoginResponseData {
  user_id: number;
  email: string;
}

export interface RegisterResponseData {
  id: number;
  email: string;
  name: string;
}

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private accessToken: string | null = null;
  private readonly api = environment.apiBaseUrl;
  private readonly loggedIn = new BehaviorSubject<boolean>(false);
  private refreshSub?: Subscription;

  constructor(private http: HttpClient) {}

  isLoggedIn(): Observable<boolean> {
    return this.loggedIn.asObservable();
  }

  get loggedInSnapshot(): boolean {
    return this.loggedIn.value;
  }

  getAccessToken(): string | null {
    return this.accessToken;
  }

  /** Bootstrap (dev bridge) or probe cookie session via GET /auth/csrf. */
  initApp(): Observable<void> {
    if (environment.useBootstrapAuth) {
      return this.loadBootstrapToken();
    }
    return this.probeCookieSession();
  }

  private loadBootstrapToken(): Observable<void> {
    return this.http
      .post<ApiEnvelope<TokenResponseData>>(
        `${this.api}/api/v1/auth/bootstrap`,
        {}
      )
      .pipe(
        tap((res) => {
          this.accessToken = res.data.access_token;
          this.loggedIn.next(true);
        }),
        map(() => undefined),
        catchError((err) => {
          this.logBootstrapError(err);
          this.loggedIn.next(false);
          this.stopSessionRefreshTimer();
          return of(undefined);
        })
      );
  }

  private probeCookieSession(): Observable<void> {
    return this.http
      .get<ApiEnvelope<{ csrf_ready?: boolean }>>(
        `${this.api}/api/v1/auth/csrf`
      )
      .pipe(
        tap(() => {
          this.loggedIn.next(true);
          this.startSessionRefreshTimer();
        }),
        map(() => undefined),
        catchError(() => {
          this.loggedIn.next(false);
          this.stopSessionRefreshTimer();
          return of(undefined);
        })
      );
  }

  /**
   * Clears local session state (no HTTP). Used when the API returns 401 on a protected call.
   */
  handleUnauthorized(): void {
    this.accessToken = null;
    this.stopSessionRefreshTimer();
    this.loggedIn.next(false);
  }

  private startSessionRefreshTimer(): void {
    this.stopSessionRefreshTimer();
    if (environment.useBootstrapAuth) {
      return;
    }
    const ms = environment.sessionRefreshIntervalMs;
    if (!ms || ms <= 0) {
      return;
    }
    this.refreshSub = interval(ms)
      .pipe(
        exhaustMap(() =>
          this.http
            .post<ApiEnvelope<{ refreshed?: boolean }>>(
              `${this.api}/api/v1/auth/refresh`,
              {}
            )
            .pipe(catchError(() => of(null)))
        )
      )
      .subscribe();
  }

  private stopSessionRefreshTimer(): void {
    this.refreshSub?.unsubscribe();
    this.refreshSub = undefined;
  }

  login(email: string, password: string): Observable<void> {
    return this.http
      .post<ApiEnvelope<LoginResponseData>>(`${this.api}/api/v1/auth/login`, {
        email,
        password
      })
      .pipe(
        tap(() => {
          this.loggedIn.next(true);
          this.startSessionRefreshTimer();
        }),
        map(() => undefined),
        catchError((err) => {
          this.logBootstrapError(err);
          throw err;
        })
      );
  }

  register(email: string, password: string, name: string): Observable<void> {
    return this.http
      .post<ApiEnvelope<RegisterResponseData>>(
        `${this.api}/api/v1/auth/register`,
        { email, password, name }
      )
      .pipe(map(() => undefined));
  }

  logout(): Observable<void> {
    if (environment.useBootstrapAuth) {
      this.accessToken = null;
      this.loggedIn.next(false);
      return of(undefined);
    }
    return this.http
      .post<ApiEnvelope<{ logged_out?: boolean }>>(
        `${this.api}/api/v1/auth/logout`,
        {}
      )
      .pipe(
        map(() => undefined),
        tap(() => {
          this.accessToken = null;
          this.loggedIn.next(false);
          this.stopSessionRefreshTimer();
        }),
        catchError(() => {
          this.accessToken = null;
          this.loggedIn.next(false);
          this.stopSessionRefreshTimer();
          return of(undefined);
        })
      );
  }

  private logBootstrapError(err: unknown): void {
    if (err instanceof HttpErrorResponse) {
      console.error(
        `AuthService: request failed ${err.status} ${err.statusText}`,
        err.error
      );
    } else {
      console.error('AuthService: request failed', err);
    }
  }
}
