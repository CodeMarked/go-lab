import { Component, OnInit } from '@angular/core';
import { PlatformService } from '../platform.service';
import {
  OperatorCaseListData,
  OperatorCaseRow,
  OperatorCaseNotesData,
  OperatorCaseActionsData
} from './cases.models';
import { SecurityMeData } from '../dataops/dataops.models';
import { MessageService } from '../message.service';

@Component({
  selector: 'app-cases',
  templateUrl: './cases.component.html',
  styleUrls: ['./cases.component.css']
})
export class CasesComponent implements OnInit {
  list: OperatorCaseListData | null = null;
  me: SecurityMeData | null = null;
  loadError = false;

  /** Create case */
  newSubjectUserId = '';
  newTitle = '';
  newDescription = '';
  newPriority = 'normal';
  createReason = '';

  /** Selected case detail */
  selected: OperatorCaseRow | null = null;
  notes: OperatorCaseNotesData | null = null;
  actions: OperatorCaseActionsData | null = null;

  patchStatus = '';
  patchPriority = '';
  patchAssigned = '';

  /** Shared audit reason for all mutations on the open case (patch, note, sanction, recovery, appeal). */
  caseActionReason = '';

  noteBody = '';

  sanctionType = 'warning';
  sanctionExpires = '';

  recoveryCharRef = '';

  appealOutcome: 'upheld' | 'overturned' = 'upheld';
  appealNotes = '';

  readonly permCasesRead = 'cases.read';
  readonly permCasesWrite = 'cases.write';
  readonly permSanctions = 'sanctions.write';
  readonly permRecovery = 'recovery.write';
  readonly permAppeals = 'appeals.resolve';

  constructor(
    private platform: PlatformService,
    private messages: MessageService
  ) {}

  ngOnInit(): void {
    this.refreshMeAndList();
  }

  refreshMeAndList(): void {
    this.loadError = false;
    this.platform.getSecurityMeTyped().subscribe({
      next: (me) => {
        this.me = me;
        this.platform.listOperatorCases({ limit: 50 }).subscribe({
          next: (data) => {
            this.list = data;
            if (!data) {
              this.loadError = true;
            }
          },
          error: () => (this.loadError = true)
        });
      },
      error: () => {
        this.me = null;
        this.loadError = true;
      }
    });
  }

  hasPerm(perm: string): boolean {
    const p = this.me?.effective_permissions;
    if (!p?.length) {
      return false;
    }
    if (p.includes('*')) {
      return true;
    }
    return p.includes(perm);
  }

  get reasonOkCreate(): boolean {
    return this.trimStr(this.createReason).length >= 10;
  }

  /** Safe trim for ngModel values that may be number or string. */
  private trimStr(v: unknown): string {
    if (v === null || v === undefined) {
      return '';
    }
    return String(v).trim();
  }

  /** Per-action check before sending the shared audit reason. */
  private confirmCaseAction(actionLabel: string): boolean {
    return window.confirm(`${actionLabel} — send the audit reason above?`);
  }

  get createCaseDisabled(): boolean {
    return !this.reasonOkCreate || !this.trimStr(this.newTitle);
  }

  submitCreate(): void {
    const uid = parseInt(this.trimStr(this.newSubjectUserId), 10);
    if (Number.isNaN(uid) || uid < 1 || !this.trimStr(this.newTitle) || !this.reasonOkCreate) {
      this.messages.add('Enter a valid subject user ID, title, and reason (at least 10 characters).');
      return;
    }
    this.platform
      .postOperatorCase(
        {
          subject_platform_user_id: uid,
          title: this.trimStr(this.newTitle),
          description: this.trimStr(this.newDescription),
          priority: this.trimStr(this.newPriority) || 'normal'
        },
        this.trimStr(this.createReason)
      )
      .subscribe((res) => {
        if (res) {
          this.messages.add(`Case ${res.id} created.`);
          this.newSubjectUserId = '';
          this.newTitle = '';
          this.newDescription = '';
          this.createReason = '';
          this.refreshMeAndList();
        } else {
          this.messages.add("Couldn't create case. See the log below for details.");
        }
      });
  }

  selectCase(row: OperatorCaseRow): void {
    this.selected = row;
    this.caseActionReason = '';
    this.patchStatus = row.status;
    this.patchPriority = row.priority;
    this.patchAssigned = row.assigned_to_user_id != null ? String(row.assigned_to_user_id) : '';
    this.notes = null;
    this.actions = null;
    this.platform.getOperatorCase(row.id).subscribe((c) => {
      if (c) {
        this.selected = c;
      }
    });
    this.platform.listOperatorCaseNotes(row.id).subscribe((d) => (this.notes = d));
    this.platform.listOperatorCaseActions(row.id).subscribe((d) => (this.actions = d));
  }

  get caseActionReasonOk(): boolean {
    return this.caseActionReason.trim().length >= 10;
  }

  submitPatch(): void {
    if (!this.selected || !this.caseActionReasonOk) {
      return;
    }
    if (!this.confirmCaseAction('Save case')) {
      return;
    }
    const body: {
      status?: string;
      priority?: string;
      assigned_to_user_id?: number | null;
    } = {};
    if (this.patchStatus.trim()) {
      body.status = this.patchStatus.trim();
    }
    if (this.patchPriority.trim()) {
      body.priority = this.patchPriority.trim();
    }
    const a = this.patchAssigned.trim();
    if (a === '') {
      body.assigned_to_user_id = null;
    } else {
      const n = parseInt(a, 10);
      if (!Number.isNaN(n) && n > 0) {
        body.assigned_to_user_id = n;
      }
    }
    this.platform.patchOperatorCase(this.selected.id, body, this.caseActionReason.trim()).subscribe((c) => {
      if (c) {
        this.messages.add(`Case ${c.id} saved.`);
        this.selected = c;
        this.caseActionReason = '';
        this.refreshMeAndList();
        this.platform.listOperatorCaseNotes(c.id).subscribe((d) => (this.notes = d));
        this.platform.listOperatorCaseActions(c.id).subscribe((d) => (this.actions = d));
      } else {
        this.messages.add("Couldn't save case. See the log below for details.");
      }
    });
  }

  get noteReasonOk(): boolean {
    return this.caseActionReasonOk && this.noteBody.trim().length > 0;
  }

  submitNote(): void {
    if (!this.selected || !this.noteReasonOk) {
      return;
    }
    if (!this.confirmCaseAction('Add note')) {
      return;
    }
    this.platform
      .postOperatorCaseNote(this.selected.id, { body: this.noteBody.trim() }, this.caseActionReason.trim())
      .subscribe((res) => {
        if (res != null && res.id != null) {
          this.messages.add('Note added.');
          this.noteBody = '';
          this.caseActionReason = '';
          this.platform.listOperatorCaseNotes(this.selected!.id).subscribe((d) => (this.notes = d));
        } else {
          this.messages.add("Couldn't add note. See the log below for details.");
        }
      });
  }

  get sanctionReasonOk(): boolean {
    return this.caseActionReasonOk && this.sanctionType.trim().length > 0;
  }

  submitSanction(): void {
    if (!this.selected || !this.sanctionReasonOk || !this.hasPerm(this.permSanctions)) {
      return;
    }
    if (!this.confirmCaseAction('Record sanction')) {
      return;
    }
    const body: { sanction_type: string; expires_at?: string } = {
      sanction_type: this.sanctionType.trim()
    };
    if (this.sanctionExpires.trim()) {
      body.expires_at = this.sanctionExpires.trim();
    }
    this.platform
      .postOperatorCaseSanction(this.selected.id, body, this.caseActionReason.trim())
      .subscribe((res) => {
        if (res?.ok) {
          this.messages.add('Sanction recorded.');
          this.caseActionReason = '';
          this.platform.listOperatorCaseActions(this.selected!.id).subscribe((d) => (this.actions = d));
        } else {
          this.messages.add("Couldn't record sanction. See the log below for details.");
        }
      });
  }

  get recoveryReasonOk(): boolean {
    return this.caseActionReasonOk;
  }

  submitRecovery(): void {
    if (!this.selected || !this.recoveryReasonOk || !this.hasPerm(this.permRecovery)) {
      return;
    }
    if (!this.confirmCaseAction('Record recovery request')) {
      return;
    }
    this.platform
      .postOperatorCaseRecoveryRequest(
        this.selected.id,
        { character_ref: this.recoveryCharRef.trim() },
        this.caseActionReason.trim()
      )
      .subscribe((res) => {
        if (res?.ok) {
          this.messages.add('Recovery request recorded.');
          this.caseActionReason = '';
          this.platform.listOperatorCaseActions(this.selected!.id).subscribe((d) => (this.actions = d));
        } else {
          this.messages.add("Couldn't record recovery request. See the log below for details.");
        }
      });
  }

  get appealReasonOk(): boolean {
    return this.caseActionReasonOk;
  }

  submitAppeal(): void {
    if (!this.selected || !this.appealReasonOk || !this.hasPerm(this.permAppeals)) {
      return;
    }
    if (!this.confirmCaseAction('Resolve appeal')) {
      return;
    }
    this.platform
      .postOperatorCaseAppealResolve(
        this.selected.id,
        { outcome: this.appealOutcome, notes: this.appealNotes.trim() },
        this.caseActionReason.trim()
      )
      .subscribe((res) => {
        if (res?.ok) {
          this.messages.add('Appeal resolved.');
          this.caseActionReason = '';
          this.platform.listOperatorCaseActions(this.selected!.id).subscribe((d) => (this.actions = d));
        } else {
          this.messages.add("Couldn't resolve appeal. See the log below for details.");
        }
      });
  }
}
