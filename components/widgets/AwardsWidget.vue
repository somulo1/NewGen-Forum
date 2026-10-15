<template>
  <div class="awards-widget">
    <div class="widget-header">
      <h3>Recent Awards</h3>
      <NuxtLink to="/awards" class="view-all">View all →</NuxtLink>
    </div>
    <div class="awards-list">
      <div v-for="award in awards" :key="award.id" class="award-item">
        <div class="award-icon">
          <Icon name="award" size="18" />
        </div>
        <div class="award-info">
          <h4>{{ award.title }}</h4>
          <p>{{ award.recipient }} • {{ award.date }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '~/components/Icon.vue'

const awards = [
  { id: 1, title: 'Excellence in Innovation', recipient: 'John Doe', date: '2 days ago' },
  { id: 2, title: 'Team Player of the Month', recipient: 'Sarah Martin', date: '1 week ago' },
  { id: 3, title: 'Outstanding Performance', recipient: 'Marcus Davis', date: '2 weeks ago' },
]
</script>

<style scoped>
.awards-widget {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
}

.widget-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-lg);
}

.widget-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
}

.view-all {
  color: var(--color-lime);
  text-decoration: none;
  font-size: 12px;
  font-weight: 600;
}

.awards-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.award-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md) 0;
  border-bottom: 1px solid var(--color-border);
}

.award-item:last-child {
  border-bottom: none;
}

.award-icon {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, var(--color-amber) 0%, var(--color-orange) 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.award-info {
  flex: 1;
}

.award-info h4 {
  margin: 0 0 var(--spacing-xs) 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.award-info p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}
</style>
