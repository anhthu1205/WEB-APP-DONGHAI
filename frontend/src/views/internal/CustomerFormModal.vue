<script setup>
import { reactive, ref } from 'vue'
import { customersApi } from '../../api/customers'
import ModalShell from '../../components/ModalShell.vue'
import PhoneInput from '../../components/PhoneInput.vue'

// Thêm khách (mọi nhân viên) hoặc sửa khách (admin)
const props = defineProps({ customer: { type: Object, default: null } })
const emit = defineEmits(['close', 'saved'])
const f = reactive({
  full_name: props.customer?.full_name || '',
  phone: props.customer?.phone || '',
  address: props.customer?.address || '',
  note: props.customer?.note || '',
})
const errs = ref({})
const busy = ref(false)
const msg = ref('')
const existingId = ref(null)

async function save() {
  busy.value = true; errs.value = {}; msg.value = ''; existingId.value = null
  const body = { full_name: f.full_name, phone: f.phone, address: f.address, note: f.note }
  try {
    const c = props.customer ? await customersApi.update(props.customer.id, body) : await customersApi.create(body)
    emit('saved', c)
  } catch (e) {
    errs.value = e.fields || {}
    msg.value = e.message
    if (e.status === 409) existingId.value = e.extra.existingId || null
  } finally { busy.value = false }
}
</script>

<template>
  <ModalShell :title="customer ? 'Sửa thông tin khách hàng' : 'Thêm khách hàng'" :busy="busy" @close="emit('close')">
    <form id="cust-form" class="form-grid" novalidate @submit.prevent="save">
      <div v-if="msg && !Object.keys(errs).length" class="alert alert-error" role="alert">
        <span>{{ msg }}<template v-if="existingId"> — <RouterLink :to="'/noi-bo/khach-hang/' + existingId" @click="emit('close')">mở khách hàng này</RouterLink></template></span>
      </div>
      <div class="field"><label for="c-name">Họ tên</label><input id="c-name" v-model="f.full_name" class="input" :class="{ invalid: errs.full_name }" maxlength="100" /><span v-if="errs.full_name" class="err">{{ errs.full_name }}</span></div>
      <div class="field"><label for="c-phone">Số điện thoại</label><PhoneInput id="c-phone" v-model="f.phone" :invalid="!!errs.phone" /><span v-if="errs.phone" class="err">{{ errs.phone }}</span></div>
      <div class="field"><label for="c-addr">Địa chỉ</label><input id="c-addr" v-model="f.address" class="input" maxlength="300" /><span v-if="errs.address" class="err">{{ errs.address }}</span></div>
      <div class="field"><label for="c-note">Ghi chú</label><textarea id="c-note" v-model="f.note" class="textarea" maxlength="1000" /></div>
    </form>
    <template #footer>
      <button class="btn btn-ghost" :disabled="busy" @click="emit('close')">Hủy</button>
      <button class="btn btn-primary" type="submit" form="cust-form" :disabled="busy"><span v-if="busy" class="spinner" />{{ customer ? 'Lưu thay đổi' : 'Thêm khách hàng' }}</button>
    </template>
  </ModalShell>
</template>
