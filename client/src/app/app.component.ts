import { Component, OnDestroy, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { Observable, of, Subscription } from 'rxjs';
import { distinctUntilChanged, switchMap } from 'rxjs/operators';
import { AuthService } from './auth.service';
import { environment } from '../environments/environment';
import { PlatformService } from './platform.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit, OnDestroy {
  title = 'Platform admin';
  loggedIn$: Observable<boolean>;
  useBootstrapAuth = environment.useBootstrapAuth;
  /** Effective platform permissions from GET /security/me (empty if not an operator). */
  effectivePermissions: string[] = [];
  private permSub?: Subscription;

  constructor(
    public auth: AuthService,
    private router: Router,
    private platform: PlatformService
  ) {
    this.loggedIn$ = this.auth.isLoggedIn();
  }

  ngOnInit(): void {
    this.permSub = this.auth
      .isLoggedIn()
      .pipe(
        distinctUntilChanged(),
        switchMap((logged) => (logged ? this.platform.getSecurityMeTyped() : of(null)))
      )
      .subscribe((me) => {
        this.effectivePermissions = me?.effective_permissions ?? [];
      });
  }

  ngOnDestroy(): void {
    this.permSub?.unsubscribe();
  }

  hasPerm(perm: string): boolean {
    const e = this.effectivePermissions;
    if (!e.length) {
      return false;
    }
    if (e.includes('*')) {
      return true;
    }
    return e.includes(perm);
  }

  logout(): void {
    this.auth.logout().subscribe(() => {
      void this.router.navigateByUrl('/login');
    });
  }
}
