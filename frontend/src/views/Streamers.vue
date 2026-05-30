<template>
  <div class="streamers-page feishu-card">
    <div class="toolbar">
      <div class="toolbar-left">
        <h2 class="page-title">主播录播管理</h2>
      </div>
      <div class="toolbar-right">
        <div class="search-box">
          <input type="text" class="feishu-input" placeholder="搜索主播昵称、ID 或平台..." />
        </div>
        <select class="feishu-input select-platform">
          <option value="">全部平台</option>
          <option value="douyin">抖音</option>
          <option value="bilibili">Bilibili</option>
        </select>
        <button class="feishu-btn">➕ 新增监听</button>
      </div>
    </div>

    <div class="table-container">
      <table class="feishu-table">
        <thead>
          <tr>
            <th width="10%">平台</th>
            <th width="25%">主播信息</th>
            <th width="30%">存储路径规则</th>
            <th width="15%">当前状态</th>
            <th width="20%">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in streamers" :key="item.id">
            <td>
              <span class="platform-tag" :class="item.platform">{{ item.platformName }}</span>
            </td>
            <td>
              <div class="streamer-name">{{ item.name }}</div>
              <div class="streamer-id">{{ item.streamId }}</div>
            </td>
            <td>
              <code class="path-code">{{ item.path }}</code>
            </td>
            <td>
              <div class="status-tag" :class="item.status">
                <span class="status-dot"></span>
                {{ item.status === 'live' ? '监听中 (直播中)' : '离线监听' }}
              </div>
            </td>
            <td>
              <button class="feishu-btn-text">编辑</button>
              <span class="divider"></span>
              <button class="feishu-btn-text text-danger">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

// 模拟前端高保真数据
const streamers = ref([
  { id: 1, platform: 'douyin', platformName: '抖音', name: '游戏区技术流', streamId: 'douyin_game_001', path: '/vol2/1000/.../抖音/game001', status: 'live' },
  { id: 2, platform: 'douyin', platformName: '抖音', name: '户外生活记录', streamId: 'dy_outdoor_xx', path: '/vol2/1000/.../抖音/outdoor', status: 'offline' },
  { id: 3, platform: 'bilibili', platformName: 'Bilibili', name: '虚拟主播VUP', streamId: 'bili_vup_123', path: '/vol2/1000/.../bilibili/vup', status: 'offline' },
  { id: 4, platform: 'douyin', platformName: '抖音', name: '美食探店', streamId: 'dy_food_888', path: '/vol2/1000/.../抖音/food', status: 'live' }
])
</script>

<style scoped>
.page-title { margin: 0; font-size: 18px; font-weight: 600; color: var(--text-title); }

/* 工具栏 */
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.toolbar-right { display: flex; gap: 12px; }
.search-box { width: 260px; }
.select-platform { width: 140px; cursor: pointer; }

/* 飞书风表格 */
.table-container { border: 1px solid var(--border-color); border-radius: var(--radius-md); overflow: hidden; }
.feishu-table { width: 100%; border-collapse: collapse; text-align: left; font-size: 14px; background: var(--bg-base); }
.feishu-table th { background: var(--bg-body); padding: 14px 16px; color: var(--text-regular); font-weight: 500; border-bottom: 1px solid var(--border-color); }
.feishu-table td { padding: 16px; border-bottom: 1px solid var(--border-color); color: var(--text-title); vertical-align: middle; }
.feishu-table tbody tr:hover { background-color: #fcfcfc; }
.feishu-table tbody tr:last-child td { border-bottom: none; }

/* 平台标签 */
.platform-tag { padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
.platform-tag.douyin { background: #1a0628; color: #ffffff; }
.platform-tag.bilibili { background: #fb7299; color: #ffffff; }

/* 文本信息 */
.streamer-name { font-weight: 500; color: var(--text-title); margin-bottom: 4px; }
.streamer-id { font-size: 12px; color: var(--text-placeholder); }
.path-code { background: var(--bg-body); padding: 4px 6px; border-radius: 4px; font-family: monospace; font-size: 12px; color: var(--text-regular); border: 1px solid var(--border-color); }

/* 状态标签 */
.status-tag { display: inline-flex; align-items: center; gap: 6px; padding: 4px 10px; border-radius: 12px; font-size: 12px; font-weight: 500; }
.status-tag.live { background: var(--color-success-bg); color: var(--color-success); }
.status-tag.offline { background: var(--bg-body); color: var(--text-regular); border: 1px solid var(--border-color); }
.status-dot { width: 6px; height: 6px; border-radius: 50%; }
.status-tag.live .status-dot { background: var(--color-success); }
.status-tag.offline .status-dot { background: var(--text-placeholder); }

/* 按钮与分隔线 */
.feishu-btn-text { background: transparent; border: none; color: var(--color-primary); cursor: pointer; font-size: 14px; padding: 4px 8px; border-radius: 4px; transition: background 0.2s; }
.feishu-btn-text:hover { background: var(--color-primary-bg); }
.feishu-btn-text.text-danger { color: var(--color-danger); }
.feishu-btn-text.text-danger:hover { background: var(--color-danger-bg); }
.divider { display: inline-block; width: 1px; height: 14px; background: var(--border-color); margin: 0 4px; vertical-align: middle; }
</style>
