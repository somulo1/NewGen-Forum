<template>
  <div class="people-container">
    <div class="page-header">
      <div>
        <h1>People</h1>
        <p class="subtitle">257 people in 11 departments across 4 branches · 152 available now</p>
      </div>
    </div>

    <div class="filters-row">
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
    </div>

    <div class="people-table">
      <div class="table-header">
        <div class="col-person">PERSON</div>
        <div class="col-department">DEPARTMENT</div>
        <div class="col-branch">BRANCH</div>
        <div class="col-status">STATUS</div>
        <div class="col-action"></div>
      </div>

      <div v-for="person in people" :key="person.id" class="table-row">
        <div class="col-person">
          <div class="person-info">
            <div class="person-avatar" :style="{ background: person.color }">{{ person.initials }}</div>
            <div class="person-details">
              <h4>{{ person.name }}</h4>
              <p>{{ person.title }}</p>
            </div>
          </div>
        </div>
        <div class="col-department">{{ person.department }}</div>
        <div class="col-branch">{{ person.branch }}</div>
        <div class="col-status">
          <span class="status-badge" :class="'status-' + person.status.toLowerCase().replace(/\s+/g, '-')">
            {{ person.status }}
          </span>
        </div>
        <div class="col-action">
          <button class="btn-message">Message</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'default',
})

const people = [
  {
    id: 1,
    initials: 'AK',
    name: 'Amina Kiprop',
    title: 'Customer Care Agent',
    color: '#7CC4FF',
    department: 'Customer Care',
    branch: 'Head Office',
    status: 'Available'
  },
  {
    id: 2,
    initials: 'SO',
    name: 'Samuel Omondi',
    title: 'Compliance Analyst',
    color: '#FFB86B',
    department: 'Risk & Compliance',
    branch: 'North Branch',
    status: 'Available'
  },
  {
    id: 3,
    initials: 'IK',
    name: 'Irene Kamau',
    title: 'Operations Officer',
    color: '#F49AC1',
    department: 'Operations',
    branch: 'Coast Branch',
    status: 'Available'
  },
  {
    id: 4,
    initials: 'AL',
    name: 'Allan Langat',
    title: 'Legal Officer',
    color: '#6EE7C8',
    department: 'Legal',
    branch: 'Head Office',
    status: 'In a meeting'
  },
  {
    id: 5,
    initials: 'JM',
    name: 'Joy Mwangi',
    title: 'Compliance Analyst',
    color: '#B8F36B',
    department: 'Risk & Compliance',
    branch: 'Head Office',
    status: 'Away'
  },
  {
    id: 6,
    initials: 'FW',
    name: 'Felix Wafula',
    title: 'Branch Supervisor',
    color: '#9B8FFF',
    department: 'Operations',
    branch: 'North Branch',
    status: 'Available'
  },
  {
    id: 7,
    initials: 'BM',
    name: 'Beatrice Muthoni',
    title: 'Legal Officer',
    color: '#FFB86B',
    department: 'Legal',
    branch: 'Head Office',
    status: 'Focusing'
  },
  {
    id: 8,
    initials: 'DO',
    name: 'Daniel Otieno',
    title: 'Recruiter',
    color: '#7CC4FF',
    department: 'HR',
    branch: 'West Branch',
    status: 'Available'
  },
  {
    id: 9,
    initials: 'AN',
    name: 'Agnes Nyambura',
    title: 'Operations Officer',
    color: '#FFB86B',
    department: 'Operations',
    branch: 'North Branch',
    status: 'Available'
  },
  {
    id: 10,
    initials: 'KW',
    name: 'Kelvin Wambui',
    title: 'Legal Officer',
    color: '#F49AC1',
    department: 'Legal',
    branch: 'Head Office',
    status: 'Available'
  },
  {
    id: 11,
    initials: 'EB',
    name: 'Esther Barasa',
    title: 'HR Officer',
    color: '#6EE7C8',
    department: 'HR',
    branch: 'West Branch',
    status: 'In a meeting'
  },
  {
    id: 12,
    initials: 'GA',
    name: 'George Akinyi',
    title: 'Credit Officer',
    color: '#B8F36B',
    department: 'Finance',
    branch: 'Head Office',
    status: 'Away'
  }
]
</script>

<style scoped>
.people-container {
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

.filters-row {
  display: grid;
  grid-template-columns: 200px 1fr 1fr 1fr;
  gap: 12px;
}

.filter-input,
.filter-select {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  padding: 8px 12px;
  font-size: 12px;
  color: var(--color-text);
  cursor: pointer;
}

.people-table {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  overflow: hidden;
}

.table-header {
  display: grid;
  grid-template-columns: 250px 180px 150px 130px 100px;
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
  grid-template-columns: 250px 180px 150px 130px 100px;
  gap: 0;
  padding: 12px 16px;
  border-bottom: 1px solid var(--color-border);
  align-items: center;
}

.table-row:hover {
  background: var(--color-bg);
}

.col-person {
  padding-right: 12px;
}

.person-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.person-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  color: white;
  flex-shrink: 0;
}

.person-details h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}

.person-details p {
  margin: 2px 0 0 0;
  font-size: 11px;
  color: var(--color-text-secondary);
}

.col-department {
  font-size: 13px;
  color: var(--color-text);
}

.col-branch {
  font-size: 13px;
  color: var(--color-text);
}

.col-status {
  font-size: 13px;
}

.status-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.status-available {
  background: rgba(184, 243, 107, 0.2);
  color: #B8F36B;
}

.status-away {
  background: rgba(155, 176, 166, 0.2);
  color: #9DB0A6;
}

.status-in-a-meeting {
  background: rgba(155, 140, 255, 0.2);
  color: #9B8FFF;
}

.status-focusing {
  background: rgba(124, 196, 255, 0.2);
  color: #7CC4FF;
}

.col-action {
  text-align: center;
}

.btn-message {
  background: transparent;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  font-size: 12px;
  font-weight: 600;
  padding: 4px 8px;
  transition: all 0.2s ease;
}

.btn-message:hover {
  color: var(--color-text);
}

@media (max-width: 1024px) {
  .filters-row {
    grid-template-columns: 1fr;
  }

  .table-header,
  .table-row {
    grid-template-columns: 1fr;
  }
}
</style>
