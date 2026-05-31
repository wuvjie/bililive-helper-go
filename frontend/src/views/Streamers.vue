<script setup>
import { useStreamerData } from '../composables/useStreamerData'

const { streamers } = useStreamerData()

function formatSize(bytes) {
  if (!bytes) return '0 MB'
  const mb = bytes / (1024 * 1024)
  if (mb > 1024) return (mb / 1024).toFixed(2) + ' GB'
  return mb.toFixed(1) + ' MB'
}

function getPlatformColor(platform) {
  if (!platform) return 'bg-gray-100 text-gray-600'
  const p = platform.toLowerCase()
  if (p.includes('douyin') || p.includes('抖音')) return 'bg-[#1a0628] text-white'
  if (p.includes('bili')) return 'bg-[#fb7299] text-white'
  return 'bg-[#e8f3ff] text-[#3370ff]'
}
</script>

<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-5">
      <h1 class="text-[17px] font-semibold text-[#1f2329]">主播管理</h1>
      <button class="px-4 py-2 bg-[#3370ff] hover:bg-[#5384ff] text-white text-sm rounded-lg transition-colors shadow-sm">
        + 新增监听
      </button>
    </div>

    <div class="bg-white border border-[#dee0e3] rounded-xl overflow-hidden shadow-sm">
      <table class="w-full text-left text-[14px]">
        <thead class="bg-[#f8f9fa] text-[#646a73] border-b border-[#dee0e3]">
          <tr>
            <th class="px-6 py-4 font-medium w-[15%]">平台</th>
            <th class="px-6 py-4 font-medium w-[30%]">主播</th>
            <th class="px-6 py-4 font-medium w-[20%]">视频大小</th>
            <th class="px-6 py-4 font-medium w-[15%]">视频数量</th>
            <th class="px-6 py-4 font-medium w-[20%]">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#dee0e3]">
          <tr v-for="s in streamers" :key="s.name" class="hover:bg-[#fcfcfc] transition-colors">
            <td class="px-6 py-4">
              <span :class="['px-2.5 py-1 rounded text-xs font-medium', getPlatformColor(s.platform)]">
                {{ s.platform || '未知平台' }}
              </span>
            </td>
            <td class="px-6 py-4">
              <div class="font-medium text-[#1f2329]">{{ s.name }}</div>
              <div class="text-xs text-[#8f959e] mt-1">{{ s.status === 'running' ? '🟢 录制中' : '⚪ 未开播' }}</div>
            </td>
            <td class="px-6 py-4 text-[#1f2329]">
              {{ formatSize(s.total_size) }}
            </td>
            <td class="px-6 py-4 text-[#1f2329]">
              {{ s.video_count || 0 }} 个
            </td>
            <td class="px-6 py-4">
              <button class="text-[#3370ff] hover:text-[#5384ff] text-sm mr-3 font-medium">编辑</button>
              <button class="text-[#f54a45] hover:text-[#ff6b6b] text-sm font-medium">删除</button>
            </td>
          </tr>
          <tr v-if="!streamers || streamers.length === 0">
            <td colspan="5" class="px-6 py-12 text-center text-[#8f959e] text-sm">
              暂无监听主播，请添加
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
