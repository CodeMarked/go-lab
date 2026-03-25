import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { HttpErrorResponse } from '@angular/common/http';
import { AuthService } from '../auth.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent implements OnInit {
  email = '';
  password = '';
  errorMsg = '';
  loading = false;
  sessionExpiredBanner = false;

  constructor(
    private auth: AuthService,
    private router: Router,
    private route: ActivatedRoute
  ) {}

  ngOnInit(): void {
    this.sessionExpiredBanner =
      this.route.snapshot.queryParamMap.get('session') === 'expired';
  }

  submit(): void {
    this.errorMsg = '';
    this.loading = true;
    this.auth.login(this.email.trim(), this.password).subscribe({
      next: () => {
        this.loading = false;
        void this.router.navigateByUrl('/dashboard');
      },
      error: (err: unknown) => {
        this.loading = false;
        if (err instanceof HttpErrorResponse) {
          const e = err.error?.error;
          this.errorMsg =
            e && typeof e.message === 'string'
              ? e.message
              : 'Login failed';
        } else {
          this.errorMsg = 'Login failed';
        }
      }
    });
  }
}
