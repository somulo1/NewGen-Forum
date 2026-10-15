<template>
  <div class="tickets-container">
    <div class="page-header">
      <div>
        <h1>Tickets</h1>
        <p class="subtitle">Raise, track and solve requests within their turnaround time.</p>
      </div>
      <button class="btn-new-ticket">New ticket</button>
    </div>

    <div class="ticket-stats">
      <div class="stat-card">
        <div class="stat-number">4</div>
        <div class="stat-label">Open</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">1</div>
        <div class="stat-label">Overdue</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">1</div>
        <div class="stat-label">Due soon</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">1</div>
        <div class="stat-label">Solved</div>
      </div>
    </div>

    <div class="integration-bar">
      <div class="integration-item">Jira · connected</div>
      <div class="integration-item">GitHub · connected</div>
      <div class="integration-item">Trello · connected</div>
      <div class="integration-item">+ Connect Asana, ClickUp...</div>
    </div>

    <div class="tickets-table">
      <div class="table-header">
        <div class="col-id">ID</div>
        <div class="col-title">TITLE</div>
        <div class="col-department">DEPARTMENT</div>
        <div class="col-from">FROM</div>
        <div class="col-category">CATEGORY</div>
        <div class="col-priority">PRIORITY</div>
        <div class="col-status">STATUS</div>
      </div>

      <div class="table-filters">
        <input type="text" class="filter-input" placeholder="Filter...">
        <input type="text" class="filter-input" placeholder="Filter...">
        <select class="filter-select">
          <option>All</option>
        </select>
        <select class="filter-select">
          <option>All</option>
        </select>
        <select class="filter-select">
          <option>All</option>
        </select>
        <select class="filter-select">
          <option>All</option>
        </select>
      </div>

      <div v-for="ticket in tickets" :key="ticket.id" class="table-row">
        <div class="col-id">{{ ticket.id }}</div>
        <div class="col-title">{{ ticket.title }}</div>
        <div class="col-department">{{ ticket.department }}</div>
        <div class="col-from">{{ ticket.from }}</div>
        <div class="col-category">{{ ticket.category }}</div>
        <div class="col-priority">
          <span class="priority-badge" :class="'priority-' + ticket.priority.toLowerCase()">
            {{ ticket.priority }}
          </span>
        </div>
        <div class="col-status">{{ ticket.status }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'default',
})

const tickets = [
  {
    id: 'IT-208',
    title: 'Laptop not connecting to office Wi-Fi',
    department: 'IT Operations',
    from: 'Customer Care',
    category: 'Network & Wi-Fi',
    priority: 'High',
    status: 'In progress'
  },
  {
    id: 'DEV-91',
    title: 'Export button fails on reports page',
    department: 'Software Development',
    from: 'HR',
    category: 'Software bug',
    priority: 'Medium',
    status: 'In progress'
  },
  {
    id: 'IT-211',
    title: 'Access request: shared finance folder',
    department: 'IT Operations',
    from: 'Sales',
    category: 'Access request',
    priority: 'Medium',
    status: 'Assigned · escalated'
  },
  {
    id: 'HR-14',
    title: 'New staff onboarding checklist update',
    department: 'HR',
    from: 'Operations',
    category: 'HR request',
    priority: 'Low',
    status: 'Resolved'
  },
  {
    id: 'DEV-95',
    title: 'Add dark mode to mobile app',
    department: 'Software Development',
    from: 'Risk & Compliance',
    category: 'Software bug',
    priority: 'Low',
    status: 'Raised'
  }
]
</script>

<style scoped>
.tickets-container {
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

.btn-new-ticket {
  background: #B8F36B;
  border: none;
  color: #000;
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.ticket-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.stat-card {
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 20px 16px;
  text-align: center;
}

.stat-number {
  font-size: 32px;
  font-weight: 700;
  color: #FFD66B;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 13px;
  color: var(--color-text-secondary);
}

.integration-bar {
  display: flex;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid var(--color-border);
}

.integration-item {
  padding: 8px 12px;
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.tickets-table {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  overflow: hidden;
}

.table-header {
  display: grid;
  grid-template-columns: 80px 250px 150px 120px 150px 100px 150px;
  gap: 0;
  background: var(--color-border);
  padding: 16px;
  font-weight: 600;
  font-size: 11px;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  border-bottom: 1px solid var(--color-border);
}

.table-filters {
  display: grid;
  grid-template-columns: 80px 250px 150px 120px 150px 100px 150px;
  gap: 0;
  padding: 12px 16px;
  border-bottom: 1px solid var(--color-border);
  align-items: center;
}

.filter-input {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 4px;
  padding: 6px 8px;
  font-size: 12px;
  color: var(--color-text);
}

.filter-select {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 4px;
  padding: 6px 8px;
  font-size: 12px;
  color: var(--color-text);
  cursor: pointer;
}

.table-row {
  display: grid;
  grid-template-columns: 80px 250px 150px 120px 150px 100px 150px;
  gap: 0;
  padding: 16px;
  border-bottom: 1px solid var(--color-border);
  align-items: center;
  font-size: 13px;
}

.table-row:hover {
  background: var(--color-bg);
}

.col-id {
  font-weight: 600;
  color: var(--color-text);
}

.col-title {
  color: var(--color-text);
  font-weight: 500;
}

.col-department {
  color: var(--color-text);
}

.col-from {
  color: var(--color-text);
}

.col-category {
  color: var(--color-text);
}

.col-priority {
  color: var(--color-text);
}

.col-status {
  color: var(--color-text-secondary);
  font-size: 12px;
}

.priority-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.priority-high {
  background: rgba(255, 155, 138, 0.2);
  color: #FF9B8A;
}

.priority-medium {
  background: rgba(255, 214, 107, 0.2);
  color: #FFD66B;
}

.priority-low {
  background: rgba(155, 176, 166, 0.2);
  color: #9DB0A6;
}

@media (max-width: 1024px) {
  .table-header,
  .table-filters,
  .table-row {
    grid-template-columns: 1fr;
  }

  .stat-card {
    grid-template-columns: 1fr;
  }

  .ticket-stats {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
