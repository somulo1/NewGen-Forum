<template>
  <div class="security-container">
    <div class="page-header">
      <div>
        <h1>Security & logs</h1>
        <p class="subtitle">Every action in Hub is logged. Rules flag anything suspicious automatically.</p>
      </div>
    </div>

    <div class="tab-controls">
      <button
        :class="['tab-btn', { active: activeTab === 'alerts' }]"
        @click="activeTab = 'alerts'"
      >
        Alerts
      </button>
      <button
        :class="['tab-btn', { active: activeTab === 'logs' }]"
        @click="activeTab = 'logs'"
      >
        Audit logs
      </button>
      <button
        :class="['tab-btn', { active: activeTab === 'rules' }]"
        @click="activeTab = 'rules'"
      >
        Detection rules
      </button>
    </div>

    <div v-if="activeTab === 'alerts'" class="alerts-tab">
      <div class="security-stats">
        <div class="stat-card">
          <div class="stat-number">6</div>
          <div class="stat-label">Open alerts</div>
          <div class="stat-detail">3 high severity</div>
        </div>
        <div class="stat-card">
          <div class="stat-number">6</div>
          <div class="stat-label">Failed sign-ins (24 h)</div>
          <div class="stat-detail">Across 1 account</div>
        </div>
        <div class="stat-card">
          <div class="stat-number">2</div>
          <div class="stat-label">Blocked attempts</div>
          <div class="stat-detail">Country and lockout rules</div>
        </div>
        <div class="stat-card">
          <div class="stat-number">186</div>
          <div class="stat-label">Active sessions</div>
          <div class="stat-detail">3 on new devices</div>
        </div>
      </div>

      <div class="activity-section">
        <h3>Activity in the last 24 hours</h3>
        <div class="activity-chart">
          <div v-for="day in 30" :key="day" class="chart-bar" :style="{ background: getBarColor(day) }"></div>
        </div>
        <div class="chart-legend">
          <span class="legend-item">
            <span class="legend-box" style="background: #FF9B8A;"></span>
            hours with flagged events
          </span>
        </div>
      </div>

      <div class="flagged-alerts-section">
        <h3>Flagged alerts</h3>
        <div class="alert-filters">
          <input type="text" class="filter-input" placeholder="Filter...">
          <select class="filter-select">
            <option>All</option>
          </select>
          <input type="text" class="filter-input" placeholder="Filter...">
          <input type="date" class="filter-input">
          <select class="filter-select">
            <option>All</option>
          </select>
        </div>

        <div class="alerts-list">
          <div v-for="alert in flaggedAlerts" :key="alert.id" class="alert-row">
            <div class="alert-id">{{ alert.id }}</div>
            <div class="alert-severity">
              <span class="severity-badge" :class="'severity-' + alert.severity.toLowerCase()">
                {{ alert.severity }}
              </span>
            </div>
            <div class="alert-title">{{ alert.title }}</div>
            <div class="alert-rule">{{ alert.rule }}</div>
            <div class="alert-date">{{ alert.date }}</div>
            <div class="alert-actions">
              <button class="action-btn">↗</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="activeTab === 'logs'" class="logs-tab">
      <div class="security-stats">
        <div class="stat-card">
          <div class="stat-number">6</div>
          <div class="stat-label">Open alerts</div>
        </div>
        <div class="stat-card">
          <div class="stat-number">6</div>
          <div class="stat-label">Failed sign-ins (24 h)</div>
        </div>
        <div class="stat-card">
          <div class="stat-number">2</div>
          <div class="stat-label">Blocked attempts</div>
        </div>
        <div class="stat-card">
          <div class="stat-number">186</div>
          <div class="stat-label">Active sessions</div>
        </div>
      </div>
      <p class="tab-placeholder">Audit logs content here</p>
    </div>

    <div v-if="activeTab === 'rules'" class="rules-tab">
      <div class="security-stats">
        <div class="stat-card">
          <div class="stat-number">6</div>
          <div class="stat-label">Open alerts</div>
        </div>
        <div class="stat-card">
          <div class="stat-number">6</div>
          <div class="stat-label">Failed sign-ins (24 h)</div>
        </div>
        <div class="stat-card">
          <div class="stat-number">2</div>
          <div class="stat-label">Blocked attempts</div>
        </div>
        <div class="stat-card">
          <div class="stat-number">186</div>
          <div class="stat-label">Active sessions</div>
        </div>
      </div>
      <p class="tab-placeholder">Detection rules content here</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

definePageMeta({
  layout: 'default',
})

const activeTab = ref('alerts')

const getBarColor = (day: number) => {
  const flaggedDays = [5, 20, 25, 26, 27, 28, 29]
  return flaggedDays.includes(day) ? '#FF9B8A' : '#6EE7C8'
}

const flaggedAlerts = [
  {
    id: 'AL-118',
    severity: 'High',
    title: '6 failed sign-ins for Peter Kamau in 5 minutes',
    rule: 'Repeated failed sign-ins',
    date: 'Today 23:28',
    description: 'From 41.90.12.7 (Nairobi, KE). The account was locked after the 6th attempt.'
  },
  {
    id: 'AL-117',
    severity: 'High',
    title: 'Diana Rotich signed in from Lagos, Nigeria for the first time',
    rule: 'Sign-in from a new country',
    date: 'Today 23:02',
    description: 'Her usual location is Nairobi, KE. IP 192.89.3.44.'
  }
]

const getAlertDetails = (id: string) => {
  return flaggedAlerts.find(a => a.id === id)
}
</script>

<style scoped>
.security-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.page-header h1 {
  margin: 0;
  font-size: 32px;
  font-weight: 600;
}

.subtitle {
  color: var(--color-text-secondary);
  font-size: 14px;
  margin: 8px 0 0 0;
}

.tab-controls {
  display: flex;
  gap: 0;
  border-bottom: 1px solid var(--color-border);
}

.tab-btn {
  background: transparent;
  border: none;
  color: var(--color-text-secondary);
  padding: 12px 16px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  border-bottom: 2px solid transparent;
  transition: all 0.2s ease;
  position: relative;
  bottom: -1px;
}

.tab-btn:hover {
  color: var(--color-text);
}

.tab-btn.active {
  color: var(--color-text);
  border-bottom-color: #B8F36B;
}

.alerts-tab,
.logs-tab,
.rules-tab {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.security-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.stat-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-number {
  font-size: 28px;
  font-weight: 700;
  color: #FF9B8A;
}

.stat-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text);
}

.stat-detail {
  font-size: 11px;
  color: var(--color-text-secondary);
}

.activity-section {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 20px;
}

.activity-section h3 {
  margin: 0 0 16px 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.activity-chart {
  display: flex;
  gap: 3px;
  margin-bottom: 12px;
  height: 60px;
}

.chart-bar {
  flex: 1;
  border-radius: 2px;
  background: #6EE7C8;
}

.chart-legend {
  font-size: 12px;
  color: var(--color-text-secondary);
  display: flex;
  gap: 8px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.legend-box {
  width: 12px;
  height: 12px;
  border-radius: 2px;
}

.flagged-alerts-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.flagged-alerts-section h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.alert-filters {
  display: grid;
  grid-template-columns: auto auto auto auto auto;
  gap: 8px;
  margin-bottom: 12px;
}

.filter-input,
.filter-select {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 4px;
  padding: 8px 12px;
  font-size: 12px;
  color: var(--color-text);
  cursor: pointer;
}

.alerts-list {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  overflow: hidden;
}

.alert-row {
  display: grid;
  grid-template-columns: 80px 80px 250px 200px 120px 60px;
  gap: 0;
  padding: 12px 16px;
  border-bottom: 1px solid var(--color-border);
  align-items: center;
  font-size: 12px;
}

.alert-row:last-child {
  border-bottom: none;
}

.alert-id {
  font-weight: 600;
  color: var(--color-text);
}

.alert-severity {
  text-align: center;
}

.severity-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.severity-high {
  background: rgba(255, 155, 138, 0.2);
  color: #FF9B8A;
}

.severity-medium {
  background: rgba(255, 214, 107, 0.2);
  color: #FFD66B;
}

.severity-low {
  background: rgba(124, 196, 255, 0.2);
  color: #7CC4FF;
}

.alert-title {
  color: var(--color-text);
  font-weight: 500;
}

.alert-rule {
  color: var(--color-text-secondary);
}

.alert-date {
  color: var(--color-text-secondary);
}

.alert-actions {
  text-align: center;
}

.action-btn {
  background: transparent;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  font-size: 16px;
}

.tab-placeholder {
  padding: 40px;
  text-align: center;
  color: var(--color-text-secondary);
}

@media (max-width: 1024px) {
  .security-stats {
    grid-template-columns: repeat(2, 1fr);
  }

  .alert-row {
    grid-template-columns: 1fr;
    gap: 8px;
  }
}
</style>
