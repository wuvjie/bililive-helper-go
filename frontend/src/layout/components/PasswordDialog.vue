<template>
  <el-dialog :model-value="visible" title="修改密码" width="400px" destroy-on-close @update:model-value="$emit('update:visible', $event)">
    <el-form label-position="top" @submit.prevent="handleSubmit">
      <el-form-item label="旧密码">
        <el-input v-model="form.old_password" type="password" show-password placeholder="输入当前密码" />
      </el-form-item>
      <el-form-item label="新密码">
        <el-input v-model="form.new_password" type="password" show-password placeholder="至少 6 个字符" />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="pw-footer">
        <el-button @click="$emit('update:visible', false)">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSubmit">确认修改</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive } from "vue";
import { ElMessage } from "element-plus";
import { changePassword } from "@/api/auth";

defineProps<{ visible: boolean }>();
const emit = defineEmits<{ "update:visible": [value: boolean] }>();

const form = reactive({ old_password: "", new_password: "" });
const saving = defineModel<boolean>("saving", { default: false });

async function handleSubmit() {
  if (!form.old_password || !form.new_password) {
    ElMessage.warning("请填写旧密码和新密码");
    return;
  }
  if (form.new_password.length < 6) {
    ElMessage.warning("新密码至少 6 个字符");
    return;
  }
  saving.value = true;
  try {
    await changePassword(form.old_password, form.new_password);
    ElMessage.success("密码已更新");
    emit("update:visible", false);
    form.old_password = "";
    form.new_password = "";
  } catch {
    /* handled by interceptor */
  } finally {
    saving.value = false;
  }
}
</script>

<style scoped>
.pw-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
