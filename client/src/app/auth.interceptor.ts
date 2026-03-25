import { Injectable } from '@angular/core';
import {
  HttpEvent,
  HttpHandler,
  HttpInterceptor,
  HttpRequest
} from '@angular/common/http';
import { Observable } from 'rxjs';
import { AuthService } from './auth.service';
import { environment } from '../environments/environment';
import { isOurApiRequest } from './api-url.util';

/** Attach Bearer JWT only when using bootstrap bridge and a token exists. */
@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  constructor(private auth: AuthService) {}

  intercept(
    req: HttpRequest<unknown>,
    next: HttpHandler
  ): Observable<HttpEvent<unknown>> {
    if (!environment.useBootstrapAuth) {
      return next.handle(req);
    }
    const token = this.auth.getAccessToken();
    const isApi = isOurApiRequest(req.url);
    if (token && isApi && !req.headers.has('Authorization')) {
      const clone = req.clone({
        setHeaders: { Authorization: `Bearer ${token}` }
      });
      return next.handle(clone);
    }
    return next.handle(req);
  }
}
