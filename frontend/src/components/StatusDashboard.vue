<script setup>
defineProps({
  diskUsage: {
    type: Number,
    default: 0
  },
  totalGB: {
    type: Number,
    default: 1
  },
  schedule: {
    type: Object,
    default: null
  },
  running: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['run'])

function formatSimpleTime(ts) {
  if (!ts || ts <= 0) return '--:--'

  const date = new Date(ts * 1000)
  const now = new Date()
  const diff = (date - now) / 1000 / 60

  const h = date.getHours().toString().padStart(2, '0')
  const m = date.getMinutes().toString().padStart(2, '0')
  const t = `${h}:${m}`

  if (Math.abs(diff) < 1) return '即将'

  if (diff < 0) {
    const ago = Math.abs(diff)
    if (ago < 60) return `${Math.round(ago)}m 前`
    return `${Math.round(ago / 60)}h 前`
  }

  return `${t} (${diff < 60 ? Math.round(diff) + 'm' : Math.round(diff / 60) + 'h'}后)`
}
</script>

<template>
  <section class="statusbar" aria-label="系统状态">
    <div class="gauge">
      <div class="gauge-val" :style="{color: diskUsage > 85 ? 'var(--err)' : diskUsage > 65 ? 'var(--warn)' : 'var(--text)'}">
        {{ diskUsage.toFixed(1) }}%
      </div>
      <div class="gauge-lbl">
        磁盘占用
        <span style="font-family:var(--font-mono);font-weight:500;color:var(--faint);margin-left:6px">
          · {{ (totalGB * (diskUsage / 100)).toFixed(1) }}G / {{ totalGB.toFixed(1) }}G
        </span>
      </div>
      <div class="gauge-track">
        <div class="gauge-fill" :style="{width: diskUsage + '%', background: diskUsage > 85 ? 'var(--err)' : diskUsage > 65 ? 'var(--warn)' : 'var(--text)'}"></div>
      </div>
    </div>

    <div class="tasks" v-if="schedule">
      <div class="task-widget">
        <div class="task-widget-head">
          <span class="task-dot" :class="schedule.merge?.enabled ? 'on' : 'off'" style="margin-right:6px"></span>
          自动合并
        </div>
        <div class="task-widget-data">
          <span class="task-widget-time" :style="{color: !schedule.merge?.enabled ? 'var(--faint)' : ''}">
            {{ schedule.merge?.enabled ? formatSimpleTime(schedule.merge?.next_run) : '-- : --' }}
          </span>
          <span class="task-widget-interval">
            {{ schedule.merge?.enabled ? schedule.merge?.interval + ' MIN' : '已暂停' }}
          </span>
        </div>
      </div>

      <div class="task-widget">
        <div class="task-widget-head">
          <span class="task-dot" :class="schedule.clean?.enabled ? 'on' : 'off'" style="margin-right:6px"></span>
          自动清理
        </div>
        <div class="task-widget-data">
          <span class="task-widget-time" :style="{color: !schedule.clean?.enabled ? 'var(--faint)' : ''}">
            {{ schedule.clean?.enabled ? formatSimpleTime(schedule.clean?.next_run) : '-- : --' }}
          </span>
          <span class="task-widget-interval">
            {{ schedule.clean?.enabled ? schedule.clean?.interval + ' MIN' : '已暂停' }}
          </span>
        </div>
      </div>
    </div>

    <div class="actions">
      <button class="btn btn-ghost auto-w" @click="emit('run', 'merge', '')" :disabled="running">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/>
          <circle cx="9" cy="7" r="4"/>
          <path d="M23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75"/>
        </svg>
        全局合并
      </button>
      <button class="btn btn-ghost auto-w" style="color:var(--err)" @click="emit('run', 'clean', '')" :disabled="running">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <polyline points="3 6 5 6 21 6"/>
          <path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
        </svg>
        全局清理
      </button>
    </div>
  </section>
</template>

<style scoped>
.statusbar {
  display: grid;
  grid-template-columns: 200px 1fr auto;
  gap: 40px;
  align-items: center;
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 24px 32px;
  margin-bottom: 24px;
  box-shadow: var(--card-shadow);
}

.gauge {
  text-align: center;
}

.gauge-val {
  font-family: var(--font-mono);
  font-size: 28px;
  font-weight: 700;
  line-height: 1;
  letter-spacing: -1px;
}

.gauge-lbl {
  font-size: 12px;
  color: var(--muted);
  font-weight: 500;
  margin-top: 6px;
}

.gauge-track {
  height: 4px;
  border-radius: 2px;
  background: var(--border-sub);
  margin-top: 10px;
  overflow: hidden;
}

.gauge-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 0.8s cubic-bezier(0.22, 1, 0.36, 1);
  background-image: linear-gradient(45deg, rgba(255,255,255,0.15) 25%, transparent 25%, transparent 50%, rgba(255,255,255,0.15) 50%, rgba(255,255,255,0.15) 75%, transparent 75%, transparent);
  background-size: 1rem 1rem;
  animation: barberpole 2s linear infinite;
}

@keyframes barberpole {
  100% { background-position: 100% 100%; }
}

.tasks {
  display: flex;
  align-items: center;
  gap: 48px;
  flex: 1;
  padding-left: 40px;
  border-left: 1px solid var(--border);
  height: 56px;
}

.task-widget {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 6px;
  min-width: 160px;
}

.task-widget-head {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 500;
  color: var(--muted);
}

.task-widget-data {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.task-widget-time {
  font-family: var(--font-mono);
  font-size: 16px;
  font-weight: 600;
  color: var(--text);
}

.task-widget-interval {
  font-size: 11px;
  font-weight: 600;
  color: var(--text2);
  background: var(--border-sub);
  padding: 2px 6px;
  border-radius: 4px;
}

.task-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.task-dot.on {
  background: var(--ok);
  box-shadow: 0 0 0 3px rgba(5,150,105,0.15);
}

.task-dot.off {
  background: var(--border);
}

.actions {
  display: flex;
  gap: 12px;
}

@media (max-width: 1024px) {
  .statusbar {
    grid-template-columns: 1fr;
    gap: 16px;
    padding: 20px;
  }

  .gauge {
    text-align: left;
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .gauge-track {
    display: none;
  }

  .tasks {
    flex-wrap: wrap;
    gap: 16px;
    padding-left: 0;
    border-left: none;
    height: auto;
    border-top: 1px solid var(--border);
    padding-top: 12px;
  }

  .task-widget {
    min-width: 0;
    flex: 1;
  }

  .actions {
    width: 100%;
    flex-wrap: wrap;
    gap: 8px;
  }

  .actions .btn {
    flex: 1;
    min-width: 0;
  }
}
</style>
