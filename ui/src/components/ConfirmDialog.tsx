import * as Dialog from "@radix-ui/react-dialog";
import Button from "@components/Button";

export default function ConfirmDialog({ open, title, description, onConfirm, onCancel }: { open: boolean; title: string; description?: string; onConfirm: () => void; onCancel: () => void }) {
  return (
    <Dialog.Root open={open}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 bg-black/20" />
        <Dialog.Content className="fixed left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-lg shadow-lg w-[360px] p-4">
          <Dialog.Title className="text-lg font-semibold mb-2">{title}</Dialog.Title>
          {description ? <Dialog.Description className="text-sm text-gray-600 mb-4">{description}</Dialog.Description> : null}
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={onCancel}>取消</Button>
            <Button variant="danger" onClick={onConfirm}>确认删除</Button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

