<script setup lang="ts">
/** Product image manager: upload, reorder (buttons + drag), delete. */
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, ChevronRight, Trash2 } from 'lucide-vue-next'
import FileDrop from '@/components/ui/FileDrop.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import { adminApi } from '@/api/admin'
import { useUiStore } from '@/stores/ui'
import { useErrorMessage } from '@/composables/useErrorMessage'
import type { ProductImage } from '@/types/api'

const images = defineModel<ProductImage[]>({ required: true })
const props = defineProps<{ productId: number }>()

const { t } = useI18n()
const ui = useUiStore()
const { errorText } = useErrorMessage()
const uploading = ref(false)
const busy = ref(false)
const toDelete = ref<ProductImage | null>(null)
const confirmOpen = ref(false)
const dragIndex = ref<number | null>(null)
const drop = ref<InstanceType<typeof FileDrop> | null>(null)

async function upload(file: File) {
  uploading.value = true
  try {
    const img = await adminApi.uploadProductImage(props.productId, file)
    images.value = [...images.value, img]
    ui.success(t('admin.common.imageUploaded'))
  } catch (e) {
    ui.error(errorText(e))
  } finally {
    uploading.value = false
    drop.value?.clearPreview()
  }
}

async function saveOrder(next: ProductImage[]) {
  const prev = images.value
  images.value = next
  busy.value = true
  try {
    await adminApi.reorderProductImages(
      props.productId,
      next.map((i) => i.id),
    )
    ui.success(t('admin.products.orderSaved'))
  } catch (e) {
    images.value = prev
    ui.error(errorText(e))
  } finally {
    busy.value = false
  }
}

function move(from: number, to: number) {
  if (to < 0 || to >= images.value.length || from === to) return
  const next = [...images.value]
  const [item] = next.splice(from, 1)
  if (!item) return
  next.splice(to, 0, item)
  void saveOrder(next)
}

function askDelete(img: ProductImage) {
  toDelete.value = img
  confirmOpen.value = true
}

async function confirmDelete() {
  const img = toDelete.value
  if (!img) return
  busy.value = true
  try {
    await adminApi.deleteProductImage(props.productId, img.id)
    images.value = images.value.filter((i) => i.id !== img.id)
    ui.success(t('admin.products.imageDeleted'))
    confirmOpen.value = false
  } catch (e) {
    ui.error(errorText(e))
  } finally {
    busy.value = false
  }
}

function onDragStart(i: number) {
  dragIndex.value = i
}
function onDrop(i: number) {
  if (dragIndex.value !== null) move(dragIndex.value, i)
  dragIndex.value = null
}
</script>

<template>
  <div class="iu">
    <ul v-if="images.length" class="list">
      <li
        v-for="(img, i) in images"
        :key="img.id"
        class="item"
        :class="{ dragging: dragIndex === i }"
        draggable="true"
        @dragstart="onDragStart(i)"
        @dragover.prevent
        @drop.prevent="onDrop(i)"
        @dragend="dragIndex = null"
      >
        <img :src="img.url" :alt="`#${i + 1}`" loading="lazy" />
        <span v-if="i === 0" class="cover">{{ $t('admin.products.cover') }}</span>
        <div class="tools">
          <button
            type="button"
            :disabled="busy || i === 0"
            :aria-label="$t('admin.products.moveLeft')"
            :title="$t('admin.products.moveLeft')"
            @click="move(i, i - 1)"
          >
            <ChevronLeft :size="16" aria-hidden="true" />
          </button>
          <button
            type="button"
            :disabled="busy || i === images.length - 1"
            :aria-label="$t('admin.products.moveRight')"
            :title="$t('admin.products.moveRight')"
            @click="move(i, i + 1)"
          >
            <ChevronRight :size="16" aria-hidden="true" />
          </button>
          <button
            type="button"
            class="del"
            :disabled="busy"
            :aria-label="$t('admin.products.deleteImage')"
            :title="$t('admin.products.deleteImage')"
            @click="askDelete(img)"
          >
            <Trash2 :size="16" aria-hidden="true" />
          </button>
        </div>
      </li>
    </ul>
    <FileDrop
      ref="drop"
      :label="$t('admin.common.uploadImage')"
      :hint="$t('admin.common.imageHint')"
      :disabled="uploading"
      @select="upload"
    />
    <ConfirmDialog
      v-model:open="confirmOpen"
      :title="$t('admin.products.deleteImage')"
      :message="$t('admin.common.confirmDeleteText')"
      :confirm-label="$t('admin.common.delete')"
      danger
      :loading="busy"
      @confirm="confirmDelete"
    />
  </div>
</template>

<style scoped>
.iu {
  display: flex;
  flex-direction: column;
  gap: var(--sp-3);
}
.list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: var(--sp-3);
}
.item {
  position: relative;
  border: 1px solid var(--c-border);
  border-radius: var(--radius-md);
  overflow: hidden;
  background: var(--c-surface-2);
  cursor: grab;
}
.item.dragging {
  opacity: 0.4;
}
.item img {
  aspect-ratio: 1;
  width: 100%;
  object-fit: cover;
}
.cover {
  position: absolute;
  top: 6px;
  left: 6px;
  background: var(--c-primary);
  color: #fff;
  font-size: 10px;
  font-weight: var(--fw-bold);
  padding: 1px 6px;
  border-radius: var(--radius-sm);
}
.tools {
  display: flex;
  justify-content: space-between;
  background: var(--c-surface);
  border-top: 1px solid var(--c-border);
}
.tools button {
  flex: 1;
  height: 34px;
  border: 0;
  background: transparent;
  cursor: pointer;
  display: grid;
  place-items: center;
  color: var(--c-text-2);
}
.tools button:hover:not(:disabled) {
  background: var(--c-surface-2);
  color: var(--c-primary);
}
.tools button.del:hover:not(:disabled) {
  color: var(--c-danger);
}
.tools button:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}
</style>
