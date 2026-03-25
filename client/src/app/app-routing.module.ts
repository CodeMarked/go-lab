import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { UsersComponent } from './users/users.component';
import { DashboardComponent } from './dashboard/dashboard.component';
import { UserDetailComponent } from './user-detail/user-detail.component';
import { LoginComponent } from './login/login.component';
import { RegisterComponent } from './register/register.component';
import { AuthGuard } from './auth.guard';
import { PlayersComponent } from './players/players.component';
import { CharactersComponent } from './characters/characters.component';
import { DataopsComponent } from './dataops/dataops.component';
import { SecurityComponent } from './security/security.component';
import { AuditComponent } from './audit/audit.component';
import { EconomyComponent } from './economy/economy.component';
import { CasesComponent } from './cases/cases.component';

const routes: Routes = [
  { path: 'login', component: LoginComponent },
  { path: 'register', component: RegisterComponent },
  {
    path: 'dashboard',
    component: DashboardComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'users',
    component: UsersComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'detail/:id',
    component: UserDetailComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'players',
    component: PlayersComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'characters',
    component: CharactersComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'economy',
    component: EconomyComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'cases',
    component: CasesComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'dataops',
    component: DataopsComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'security',
    component: SecurityComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'audit',
    component: AuditComponent,
    canActivate: [AuthGuard]
  },
  { path: '', redirectTo: '/dashboard', pathMatch: 'full' },
  { path: '**', redirectTo: '/dashboard' }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule {}
