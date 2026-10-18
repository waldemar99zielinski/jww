import { IconClose, IconEdit } from '@/components/icons'
import { IconButton } from '@/components/ui/IconButton'
import { Loader } from '@/components/ui/Loader'
import { AttachmentList } from './components/AttachmentList'
import { DetailList } from './components/DetailList'
import { SwipeToClose } from './components/SwipeToClose'
import { TransactionEditForm } from './components/TransactionEditForm'
import { TransactionHero } from './components/TransactionHero'
import type { TransactionDetailViewProps } from './types'

export default function TransactionDetailMobile({
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
    <SwipeToClose onClose={onClose}>
      {(close) => (
        <div className="flex flex-col gap-4 px-4 pb-4">
          <div className="sticky top-0 z-20 -mx-4 flex items-center justify-between bg-paper px-4 pb-3 pt-6">
            <h2 className="m-0 text-[19px] font-bold">{isEditing ? 'Edit transaction' : 'Transaction'}</h2>
            <div className="flex items-center gap-2">
              {!isLoading && !isError && !isEditing && (
                <IconButton
                  aria-label="Edit"
                  onClick={onStartEdit}
                  icon={<IconEdit className="h-4 w-4" />}
                  className="h-9 w-9"
                />
              )}
              <IconButton aria-label="Close" onClick={close} icon={<IconClose className="h-4 w-4" />} className="h-9 w-9" />
            </div>
          </div>

          {isLoading && (
            <div className="flex justify-center py-10">
              <Loader />
            </div>
          )}

          {isError && !isLoading && (
            <p className="m-0 text-xs text-muted">Couldn't load this transaction. It may have been deleted.</p>
          )}

          {!isLoading && !isError && (
            <>
              <div className="flex flex-col gap-4 rounded-2xl border border-line bg-surface p-5">
                {isEditing ? (
                  <TransactionEditForm {...edit} layout="mobile" />
                ) : (
                  <>
                    <TransactionHero title={title} categoryName={categoryName} amount={amount} isIncome={isIncome} />
                    <DetailList items={items} />
                  </>
                )}
              </div>

              {attachments.length > 0 && (
                <section className="flex flex-col gap-2">
                  <h2 className="m-0 text-[11.5px] font-bold uppercase tracking-wide text-muted">Attachments</h2>
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
      )}
    </SwipeToClose>
  )
}
