<script setup>
import { useStreamerData } from '../composables/useStreamerData'

const { streamers } = useStreamerData()

function formatSize(gb) {
  if (!gb && gb !== 0) return '-'
  if (gb < 1) return (gb * 1024).toFixed(1) + ' MB'
  return gb.toFixed(2) + ' GB'
}
</script>

<template>
  <div class="p-6" style="padding: 32px 40px;">
    <div class="flex items-center justify-between mb-5">
      <h2 class="section-title" style="margin-bottom:0;">主播管理</h2>
      <button class="feishu-btn" @click="$router.push('/tasks')">
        ⚙️ 任务中心
      </button>
    </div>

    <div class="bg-white border border-[#dee0e3] rounded-xl overflow-hidden shadow-sm">
      <table class="w-full text-left text-[14px]">
        <thead class="bg-[#f8f9fa] text-[#646a73] border-b border-[#dee0e3]">
          <tr>
            <th class="px-6 py-4 font-medium w-[40%]">主播</th>
            <th class="px-6 py-4 font-medium w-[30%]">录播文件数</th>
            <th class="px-6 py-4 font-medium w-[30%]">占用空间</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#dee0e3]">
          <tr v-for="s in streamers" :key="s.name" class="hover:bg-[#fcfcfc] transition-colors">
            <td class="px-6 py-4">
              <div class="font-medium text-[#1f2329]">{{ s.name }}</div>
            </td>
            <td class="px-6 py-4 text-[#1f2329]">
              {{ s.files || 0 }} 个文件
            </td>
            <td class="px-6 py-4 text-[#1f2329]">
              {{ formatSize(s.size_gb) }}
            </td>
          </tr>
          <tr v-if="!streamers || streamers.length === 0">
            <td colspan="3" class="px-6 py-16 text-center text-[#8f959e]">
              <div class="text-4xl mb-3 opacity-50">📂</div>
              暂无录播数据
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
