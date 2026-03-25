import { Injectable, Injector } from '@angular/core';
import {
  HttpErrorResponse,
  HttpEvent,
  HttpHandler,
  HttpInterceptor,
  HttpRequest
} from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { Router } from '@angular/router';
import { AuthService } from './auth.service';
import { isOurApiRequest } from './api-url.util';

/** Paths where 401 is an expected client outcome (not “signed out elsewhere”). */
function ignore401Navigation(req: HttpRequest<unknown>): boolean {
  if (!isOurApiRequest(req.url)) {
    return true;
  }
  const u = req.url;
  const m = req.method.toUpperCase();
  if (m === 'GET' && u.includes('/api/v1/auth/csrf')) {
    return true;
  }
  const authExempt = [
    '/api/v1/auth/login',
    '/api/v1/auth/register',
    '/api/v1/auth/bootstrap',
    '/api/v1/auth/token'
  ];
  return authExempt.some((p) => u.includes(p));
}

@Injectable()
export class UnauthorizedInterceptor implements HttpInterceptor {
  constructor(private injector: Injector) {}

  intercept(
    req: HttpRequest<unknown>,
    next: HttpHandler
  ): Observable<HttpEvent<unknown>> {
    return next.handle(req).pipe(
      catchError((err: unknown) => {
        if (
          err instanceof HttpErrorResponse &&
          err.status === 401 &&
          !ignore401Navigation(req)
        ) {
          const auth = this.injector.get(AuthService);
          auth.handleUnauthorized();
          const router = this.injector.get(Router);
          queueMicrotask(() =>
            router.navigate(['/login'], { queryParams: { session: 'expired' } })
          );
        }
        return throwError(() => err);
      })
    );
  }
}
