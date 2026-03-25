import { Component, EventEmitter, Input, Output } from '@angular/core';

/** Matches `RequirePlatformActionReason` default in the API. */
export const MIN_PLATFORM_ACTION_REASON_LENGTH = 10;

@Component({
  selector: 'app-platform-action-reason',
  templateUrl: './platform-action-reason.component.html',
  styleUrls: ['./platform-action-reason.component.css']
})
export class PlatformActionReasonComponent {
  @Input() reason = '';
  @Output() reasonChange = new EventEmitter<string>();

  @Input() minLength = MIN_PLATFORM_ACTION_REASON_LENGTH;
  @Input() name = 'platform-action-reason';
  @Input() label = 'Audit reason';
  /** Shown when the value is non-empty but shorter than minLength. */
  @Input() invalidHint = '';
  @Input() placeholder = '';
  @Input() rows = 2;
  /** One line tying the field to the platform header (hide for very compact layouts). */
  @Input() showHeaderLine = true;

  onReasonInput(v: string): void {
    this.reason = v;
    this.reasonChange.emit(v);
  }

  /** Use from parent templates as `#ref.ok` for button disabled state. */
  get ok(): boolean {
    return this.reason.trim().length >= this.minLength;
  }

  get showWarn(): boolean {
    return !this.ok && this.reason.length > 0;
  }
}
