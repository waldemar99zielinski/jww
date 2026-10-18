import { IconClose, IconEdit } from '@/components/icons'
import { IconButton } from '@/components/ui/IconButton'
import { Loader } from '@/components/ui/Loader'
import { AttachmentList } from './components/AttachmentList'
import { DetailList } from './components/DetailList'
import { TransactionEditForm } from './components/TransactionEditForm'
import { TransactionHero } from './components/TransactionHero'
import type { TransactionDetailViewProps } from './types'

export default function TransactionDetailDesktop({
  title,
  categoryName,
  amount,
  isIncome,
  items,
  attachments,
  downloadError,
  isLoading,
  isError,
  isEditing,
  onStartEdit,
  edit,
  onPreview,
  onOpenActions,
  onClose,
}: TransactionDetailViewProps) {
  return (
    <div className="screen-enter mx-auto flex max-w-2xl flex-col gap-6 rounded-2xl border border-line bg-surface p-8">
      <div className="flex items-center justify-between">
        <h1 className="m-0 text-2xl font-bold">{isEditing ? 'Edit transaction' : 'Transaction'}</h1>
        <div className="flex items-center gap-2">
          {!isLoading && !isError && !isEditing && (
            <IconButton aria-label="Edit" onClick={onStartEdit} icon={<IconEdit className="h-4 w-4" />} className="h-9 w-9" />
          )}
          <IconButton aria-label="Close" onClick={onClose} icon={<IconClose className="h-4 w-4" />} className="h-9 w-9" />
        </div>
      </div>

      {isLoading && (
        <div className="flex justify-center py-10">
          <Loader />
        </div>
      )}

      {isError && !isLoading && (
        <p className="m-0 text-sm text-muted">Couldn't load this transaction. It may have been deleted.</p>
      )}

      {!isLoading && !isError && (
        <>
          {isEditing ? (
            <TransactionEditForm {...edit} layout="desktop" />
          ) : (
            <>
              <TransactionHero title={title} categoryName={categoryName} amount={amount} isIncome={isIncome} />
              <DetailList items={items} />
            </>
          )}
          {attachments.length > 0 && (
            <section className="flex flex-col gap-2">
              <h2 className="m-0 text-xs font-semibold uppercase tracking-wide text-muted">Attachments</h2>
              <AttachmentList
                attachments={attachments}
                downloadError={downloadError}
                onPreview={onPreview}
                onOpenActions={onOpenActions}
              />
            </section>
          )}
        </>
      )}
    </div>
  )
}
