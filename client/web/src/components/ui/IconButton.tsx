import type { ReactNode } from 'react'
import { Button, type ButtonProps } from './Button'
import { cx } from './classes'

export interface IconButtonProps extends ButtonProps {
  icon: ReactNode
}

export function IconButton({ icon, className, variant = 'custom', children, ...props }: IconButtonProps) {
  return (
    <Button
      variant={variant}
      className={cx('flex items-center justify-center rounded-[10px] border border-line bg-surface-sunken', className)}
      {...props}
    >
      {icon}
      {children}
    </Button>
  )
}
