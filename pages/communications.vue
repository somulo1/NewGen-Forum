<template>
  <div class="communications-container">
    <div class="page-header">
      <div>
        <h1>Communications</h1>
        <p class="subtitle">Manage communication rules and risks across your organization.</p>
      </div>
    </div>

    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-number">Risk & Compliance</div>
        <div class="stat-label">My department</div>
        <div class="stat-detail">1 branch unit</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">18</div>
        <div class="stat-label">Communication rules</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">3</div>
        <div class="stat-label">Risks identified</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">67%</div>
        <div class="stat-label">Resolution rate</div>
        <div class="progress-bar">
          <div class="progress-fill" style="width: 67%"></div>
        </div>
      </div>
    </div>

    <div class="communications-table">
      <div class="table-header">
        <div class="col-comm">COMMUNICATION</div>
        <div class="col-rule">RULE</div>
        <div class="col-status">STATUS</div>
      </div>

      <div v-for="comm in communications" :key="comm.id" class="table-row">
        <div class="col-comm">{{ comm.name }}</div>
        <div class="col-rule">{{ comm.rule }}</div>
        <div class="col-status">
          <span class="status-badge" :class="'status-' + comm.status.toLowerCase()">{{ comm.status }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'default',
})

const communications = [
  { id: 1, name: 'Email governance', rule: 'Encrypted for external', status: 'Active' },
  { id: 2, name: 'File sharing', rule: 'Allowed to branch level only', status: 'Active' },
  { id: 3, name: 'External messaging', rule: 'Requires approval', status: 'Active' },
  { id: 4, name: 'Third-party access', rule: 'Time-limited tokens', status: 'Pending review' },
]
</script>

<style scoped>
.communications-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
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

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.stat-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 16px;
}

.stat-number {
  font-size: 18px;
  font-weight: 700;
  color: var(--color-text);
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-secondary);
  margin-bottom: 4px;
}

.stat-detail {
  font-size: 11px;
  color: var(--color-text-secondary);
}

.progress-bar {
  height: 4px;
  background: var(--color-border);
  border-radius: 2px;
  overflow: hidden;
  margin-top: 8px;
}

.progress-fill {
  height: 100%;
  background: #B8F36B;
  border-radius: 2px;
}

.communications-table {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  overflow: hidden;
}

.table-header {
  display: grid;
  grid-template-columns: 1fr 1fr 150px;
  gap: 0;
  background: var(--color-border);
  padding: 12px 16px;
  font-weight: 600;
  font-size: 11px;
  color: var(--color-text-secondary);
  text-transform: uppercase;
}

.table-row {
  display: grid;
  grid-template-columns: 1fr 1fr 150px;
  gap: 0;
  padding: 12px 16px;
  border-bottom: 1px solid var(--color-border);
  align-items: center;
  font-size: 13px;
}

.table-row:hover {
  background: var(--color-bg);
}

.status-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.status-active {
  background: rgba(184, 243, 107, 0.2);
  color: #B8F36B;
}

.status-pending {
  background: rgba(255, 214, 107, 0.2);
  color: #FFD66B;
}

@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .table-header,
  .table-row {
    grid-template-columns: 1fr;
  }
}
</style>
