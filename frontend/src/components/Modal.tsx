import type { ReactNode, MouseEvent } from 'react'
import './Modal.css'

interface ModalProps {
  open: boolean
  onClose: () => void
  title: string
  subtitle?: string
  children: ReactNode
  footer: ReactNode
  /** Widens the card for content that doesn't fit the default 440px form
   *  width — a sortable table, say — without forcing every other modal to
   *  carry the option. */
  wide?: boolean
}

export function Modal({ open, onClose, title, subtitle, children, footer, wide = false }: ModalProps) {
  if (!open) return null

  const stop = (e: MouseEvent) => e.stopPropagation()

  return (
    <div className="modal-scrim" onClick={onClose}>
      <div className={'modal-card' + (wide ? ' modal-card-wide' : '')} onClick={stop}>
        <div className="modal-header">
          <div className="modal-title">{title}</div>
          {subtitle && <div className="modal-subtitle">{subtitle}</div>}
        </div>
        <div className="modal-body">{children}</div>
        <div className="modal-footer">{footer}</div>
      </div>
    </div>
  )
}

export function ModalCancelButton({ onClick }: { onClick: () => void }) {
  return (
    <button type="button" className="btn btn-secondary" onClick={onClick}>
      Cancel
    </button>
  )
}
