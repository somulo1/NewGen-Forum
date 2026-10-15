<template>
  <div class="approvals-container">
    <div class="page-header">
      <div>
        <h1>Approvals</h1>
        <p class="subtitle">Requests move through each approval level in order. Some need one approver, some two or three.</p>
      </div>
      <div class="status-filter">
        <button class="filter-btn active">Waiting ({{ waitingCount }})</button>
        <button class="filter-btn">All</button>
      </div>
    </div>

    <div class="approvals-list">
      <div v-for="approval in approvals" :key="approval.id" class="approval-card">
        <div class="approval-header">
          <div class="approval-id">{{ approval.id }}</div>
          <div class="approval-category">{{ approval.category }}</div>
          <h3 class="approval-title">{{ approval.title }}</h3>
          <span class="approval-level">Level {{ approval.level }} of {{ approval.maxLevel }}</span>
        </div>

        <div class="approval-meta">
          <p>Requested by {{ approval.requester }} · {{ approval.time }} · {{ approval.detail }}</p>
        </div>

        <div class="approval-content">
          <p>{{ approval.description }}</p>
        </div>

        <div class="approval-workflow">
          <div v-for="(step, idx) in approval.workflow" :key="idx" class="workflow-step">
            <div class="step-icon" :class="'step-' + step.status">
              <span v-if="step.status === 'approved'">✓</span>
              <span v-else-if="step.status === 'waiting'">⭕</span>
              <span v-else>○</span>
            </div>
            <div class="step-info">
              <h4>{{ step.role }}</h4>
              <p v-if="step.status === 'approved'">Approved by {{ step.approver }} · {{ step.time }}</p>
              <p v-else-if="step.status === 'waiting'">Waiting now</p>
              <p v-else>{{ step.approver }}</p>
            </div>
          </div>
        </div>

        <div class="approval-actions">
          <p class="waiting-message">{{ approval.waitingMessage }}</p>
          <div class="action-buttons">
            <button class="btn-reject">Reject</button>
            <button class="btn-approve">Approve</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'default',
})

const waitingCount = ref(3)

const approvals = [
  {
    id: 'AP-33',
    category: 'Files',
    title: 'Share "Q3 budget tracker.xlsx" with the external auditor',
    level: 1,
    maxLevel: 1,
    requester: 'Amina Wanjiru',
    time: 'Today, 11:40',
    detail: 'Direct approval',
    description: 'The auditor needs it before Friday\'s review.',
    workflow: [
      {
        role: 'Line manager',
        status: 'waiting',
        approver: '',
        time: '',
      },
    ],
    waitingMessage: 'Waiting for you as Line manager',
    actionBtnColor: '#FFB86B',
  },
  {
    id: 'AP-32',
    category: 'Announcements',
    title: 'Announcement: Office closed Friday for maintenance',
    level: 2,
    maxLevel: 2,
    requester: 'Faith Achieeng',
    time: 'Today, 09:15',
    detail: '2-level approval',
    description: '',
    workflow: [
      {
        role: 'Head of department',
        status: 'approved',
        approver: 'Grace Njeri',
        time: 'Today, 10:02',
      },
      {
        role: 'Communications lead',
        status: 'waiting',
        approver: '',
        time: '',
      },
    ],
    waitingMessage: 'Waiting for you as Communications lead',
    actionBtnColor: '#FFB86B',
  },
  {
    id: 'AP-31',
    category: 'Projects',
    title: 'New project: Customer self-service portal',
    level: 3,
    maxLevel: 3,
    requester: 'Kevin Mutua',
    time: 'Yesterday, 16:20',
    detail: '3-level approval',
    description: 'Estimated budget [AMOUNT]. Six months.',
    workflow: [
      {
        role: 'Head of department',
        status: 'approved',
        approver: 'Grace Njeri',
        time: 'Yesterday, 17:00',
      },
      {
        role: 'Finance approver',
        status: 'approved',
        approver: 'Amina Wanjiru',
        time: 'Today, 08:30',
      },
      {
        role: 'Managing Director',
        status: 'waiting',
        approver: '',
        time: '',
      },
    ],
    waitingMessage: 'Waiting for you as Managing Director',
    actionBtnColor: '#FFB86B',
  },
]
</script>

<style scoped>
.approvals-container {
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

.status-filter {
  display: flex;
  gap: 8px;
}

.filter-btn {
  background: transparent;
  border: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.filter-btn.active {
  background: var(--color-surface);
  border-color: var(--color-text);
  color: var(--color-text);
}

.filter-btn:hover {
  border-color: #B8F36B;
  color: var(--color-text);
}

.approvals-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.approval-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.approval-header {
  display: flex;
  align-items: baseline;
  gap: 16px;
  flex-wrap: wrap;
}

.approval-id {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.approval-category {
  display: inline-block;
  background: var(--color-border);
  color: var(--color-text-secondary);
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.approval-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
  flex: 1;
}

.approval-level {
  background: #FFB86B;
  color: #000;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  white-space: nowrap;
}

.approval-meta {
  padding: 12px 0;
  border-top: 1px solid var(--color-border);
  border-bottom: 1px solid var(--color-border);
}

.approval-meta p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.approval-content {
  padding: 12px 0;
}

.approval-content p {
  margin: 0;
  font-size: 13px;
  color: var(--color-text);
  line-height: 1.5;
}

.approval-workflow {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px 0;
  border-top: 1px solid var(--color-border);
  border-bottom: 1px solid var(--color-border);
}

.workflow-step {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.step-icon {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  flex-shrink: 0;
  border: 2px solid;
}

.step-icon.step-approved {
  background: #B8F36B;
  border-color: #B8F36B;
  color: #000;
}

.step-icon.step-waiting {
  background: transparent;
  border-color: #FFB86B;
  color: #FFB86B;
}

.step-icon.step-pending {
  background: transparent;
  border-color: var(--color-border);
  color: var(--color-text-secondary);
}

.step-info h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}

.step-info p {
  margin: 2px 0 0 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.approval-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
}

.waiting-message {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.action-buttons {
  display: flex;
  gap: 12px;
}

.btn-reject {
  background: transparent;
  border: 1px solid #FF9B8A;
  color: #FF9B8A;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-reject:hover {
  background: rgba(255, 155, 138, 0.1);
}

.btn-approve {
  background: #B8F36B;
  border: none;
  color: #000;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s ease;
}

.btn-approve:hover {
  opacity: 0.9;
}

@media (max-width: 768px) {
  .approval-header {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .approval-level {
    align-self: flex-start;
  }
  
  .approval-actions {
    flex-direction: column;
    gap: 16px;
  }
  
  .action-buttons {
    width: 100%;
  }
  
  .btn-reject,
  .btn-approve {
    flex: 1;
  }
}
</style>
