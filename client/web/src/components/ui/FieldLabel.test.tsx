import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { FieldLabel } from '@/components/ui/FieldLabel'

describe('FieldLabel', () => {
  it('associates the label text with its child control via implicit <label> nesting', () => {
    render(
      <FieldLabel label="Name">
        <input />
      </FieldLabel>,
    )

    expect(screen.getByLabelText('Name')).toBeInTheDocument()
  })
})
