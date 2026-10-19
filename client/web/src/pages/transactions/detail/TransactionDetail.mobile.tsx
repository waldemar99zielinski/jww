import { IconClose, IconEdit, IconPaperclip, IconTrash } from '@/components/icons'
import { Button } from '@/components/ui/Button'
import { IconButton } from '@/components/ui/IconButton'
import { Loader } from '@/components/ui/Loader'
import { AttachmentList } from './components/AttachmentList'
import { DetailList } from './components/DetailList'
import { SubTransactionSummary } from './components/SubTransactionSummary'
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
  subTransactions,
  attachments,
  attachmentsUploading,
  attachmentUploadError,
  downloadError,
  isLoading,
  isError,
  isEditing,
  onStartEdit,
  onAddAttachment,
  onRequestDelete,
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
                    {subTransactions.length > 0 && <SubTransactionSummary items={subTransactions} />}
                  </>
                )}
              </div>

              <section className="flex flex-col gap-2">
                <div className="flex items-center justify-between gap-3">
                  <h2 className="m-0 text-[11.5px] font-bold uppercase tracking-wide text-muted">Attachments</h2>
                  <Button
                    variant="link"
                    onClick={onAddAttachment}
                    disabled={attachmentsUploading}
                    className="flex items-center gap-1"
                  >
                    <IconPaperclip className="h-3.5 w-3.5" />
                    {attachmentsUploading ? 'Uploading…' : 'Add attachment'}
                  </Button>
                </div>
                {attachments.length > 0 ? (
                  <AttachmentList
                    attachments={attachments}
                    downloadError={downloadError}
                    onPreview={onPreview}
                    onOpenActions={onOpenActions}
                  />
                ) : (
                  <p className="m-0 text-xs text-muted">No attachments yet.</p>
                )}
                {attachmentUploadError && (
                  <p className="m-0 text-[11.5px] text-critical">Couldn't upload that file — try again.</p>
                )}
              </section>

              {!isEditing && (
                <Button
                  variant="danger"
                  fullWidth
                  onClick={onRequestDelete}
                  className="flex items-center justify-center gap-2"
                >
                  <IconTrash className="h-4 w-4" />
                  Remove transaction
                </Button>
              )}
            </>
          )}
        </div>
      )}
    </SwipeToClose>
  )
}
