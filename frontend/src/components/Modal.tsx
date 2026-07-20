import type { ReactNode, MouseEvent } from 'react'
import './Modal.css'

interface ModalProps {
  open: boolean
  onClose: () => void
  title: string
  subtitle?: string
  children: ReactNode
  footer: ReactNode
}

export function Modal({ open, onClose, title, subtitle, children, footer }: ModalProps) {
  if (!open) return null

  const stop = (e: MouseEvent) => e.stopPropagation()

  return (
    <div className="modal-scrim" onClick={onClose}>
      <div className="modal-card" onClick={stop}>
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
