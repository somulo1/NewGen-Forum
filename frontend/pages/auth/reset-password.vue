<script setup>
// Reset password, in four steps, all on the server: 1 email -> 2 the 6-digit code that was emailed -> 3 the three security
// questions -> 4 a link is emailed; open it, enter its code and choose the new password. The page says the same thing whether or
// not an account exists.
definePageMeta({ path: '/reset-password', layout: 'auth' });
useHead({ title: 'Reset password · Hub' });
const route = useRoute();
const be = useBackend();
const KEY = 'hub-reset';

const st = reactive({ stage: 'email', email: '', resetId: '', questions: [] });
const load = () => { try { Object.assign(st, JSON.parse(sessionStorage.getItem(KEY) || '{}')); } catch (e) { /* start fresh */ } };
const save = () => { try { sessionStorage.setItem(KEY, JSON.stringify(st)); } catch (e) { /* storage blocked */ } };
const reset = (msg = '') => { Object.assign(st, { stage: 'email', email: '', resetId: '', questions: [] }); notice.value = msg; err.value = ''; code.value = ''; try { sessionStorage.removeItem(KEY); } catch (e) { /* nothing to clear */ } };

const email = ref(''), code = ref(''), answers = ref(['', '', '']), pw = ref(''), pw2 = ref(''), showPw = ref(false);
const err = ref(''), notice = ref(''), busy = ref(false), cool = ref(0);
let timer = null;
const startCool = () => { cool.value = 30; clearInterval(timer); timer = setInterval(() => { cool.value--; if (cool.value <= 0) clearInterval(timer); }, 1000); };
const msg = (e, fallback) => (e && e.message) || fallback;

// Opened from the emailed link (?token=...): choose the new password with the code from that email.
const openLink = () => { const t = route.query.token; if (!t) return; st.stage = 'link'; code.value = ''; err.value = ''; save(); };
onMounted(() => { load(); openLink(); });
watch(() => route.query.token, () => openLink());
onBeforeUnmount(() => clearInterval(timer));

// 1. Email.
const start = async (again) => {
  err.value = ''; notice.value = '';
  const e = (again ? st.email : email.value).trim().toLowerCase();
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(e)) { err.value = 'Enter your email.'; return; }
  busy.value = true;
  try {
    const r = await be.open('POST', '/auth/reset/start', { email: e });
    Object.assign(st, { email: e, resetId: r.reset_id, stage: 'otp' });
    code.value = ''; startCool(); save();
    if (again) notice.value = 'Code sent.';
  } catch (x) { err.value = msg(x, 'Could not start the reset. Try again.'); }
  busy.value = false;
};
const sendCode = () => start(false);
const resend = () => { if (cool.value <= 0) start(true); };

// 2. Code from the email.
const checkCode = async () => {
  err.value = ''; notice.value = '';
  busy.value = true;
  try {
    const r = await be.open('POST', '/auth/reset/code', { reset_id: st.resetId, code: code.value.trim() });
    st.questions = r.questions || []; st.stage = 'quiz'; answers.value = st.questions.map(() => ''); save();
  } catch (x) { err.value = msg(x, 'Incorrect code.'); }
  busy.value = false;
};

// 3. Security questions. All answers must be right.
const questions = computed(() => st.questions);
const checkQuiz = async () => {
  err.value = '';
  if (answers.value.some((a) => a.trim().length < 1)) { err.value = 'Answer all questions.'; return; }
  busy.value = true;
  try {
    await be.open('POST', '/auth/reset/quiz', { reset_id: st.resetId, answers: answers.value });
    st.stage = 'sent'; save();
  } catch (x) { err.value = msg(x, 'Incorrect answers.'); }
  busy.value = false;
};

// 4. The link from the second email, its code, and the new password.
const rules = computed(() => [['At least 8 characters', pw.value.length >= 8], ['An uppercase and a lowercase letter', /[a-z]/.test(pw.value) && /[A-Z]/.test(pw.value)], ['A number', /\d/.test(pw.value)]]);
const setNew = async () => {
  err.value = '';
  if (rules.value.some((r) => !r[1])) { err.value = 'Password does not meet the rules.'; return; }
  if (pw.value !== pw2.value) { err.value = 'Passwords do not match.'; return; }
  busy.value = true;
  try {
    await be.open('POST', '/auth/reset/complete', { token: String(route.query.token || ''), code: code.value.trim(), new_password: pw.value });
  } catch (x) {
    err.value = msg(x, 'Could not reset your password.');
    busy.value = false;
    return;
  }
  busy.value = false;
  try { sessionStorage.removeItem(KEY); } catch (e) { /* nothing to clear */ }
  navigateTo('/login?reset=1');
};

const field = 'height: 44px; padding: 0 14px; border-radius: 12px; border: 1px solid var(--k-C9D6CF,#C9D6CF); background: #fff; color: var(--k-0F1714,#0F1714); font-size: 15px';
const primary = 'height: 46px; border-radius: 23px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 700; font-size: 15px; cursor: pointer';
const label = 'display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-55645C,#55645C)';
const head = "margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 26px; letter-spacing: -0.02em";
const steps = ['1', '2', '3', '4'];
const stepNo = computed(() => ({ email: 0, otp: 1, quiz: 2, sent: 3, link: 3 })[st.stage]);
</script>
<template>
  <div style="display: flex; flex-direction: column; gap: 14px">
    <ol aria-label="Progress" style="list-style: none; margin: 0; padding: 0; display: flex; gap: 6px">
      <li v-for="(s, i) in steps" :key="s" :aria-current="i === stepNo ? 'step' : undefined" style="flex: 1; display: flex; flex-direction: column; gap: 4px; min-width: 0"><span :style="`height: 4px; border-radius: 2px; background: ${i <= stepNo ? '#7BC51F' : 'var(--k-DCE5DF,#DCE5DF)'}`"></span><span :style="`font-size: 10px; text-align: center; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: ${i === stepNo ? 'var(--k-0F1714,#0F1714)' : 'var(--k-8A988F,#8A988F)'}; font-weight: ${i === stepNo ? 700 : 500}`">{{ s }}</span></li>
    </ol>
    <div v-if="notice" role="status" style="padding: 10px 12px; border-radius: 10px; background: #EAF6DD; color: #2B5A0B; font-size: 13px">{{ notice }}</div>

    <form v-if="st.stage === 'email'" novalidate aria-label="Reset password" @submit.prevent="sendCode" style="display: flex; flex-direction: column; gap: 14px">
      <h1 :style="head">Reset password</h1>
      <label :style="label">Email<input v-model="email" type="email" autocomplete="email" placeholder="you@company.com" :style="field"></label>
      <span v-if="err" role="alert" style="color: #C0392B; font-size: 13px">{{ err }}</span>
      <button type="submit" :disabled="busy" :style="primary">Continue</button>
    </form>

    <form v-else-if="st.stage === 'otp'" novalidate aria-label="Enter the code" @submit.prevent="checkCode" style="display: flex; flex-direction: column; gap: 14px">
      <h1 :style="head">Enter code</h1>
      <label :style="label">Code<input v-model="code" type="text" inputmode="numeric" autocomplete="one-time-code" maxlength="6" placeholder="••••••" aria-label="Code" :style="field + '; letter-spacing: 0.4em; text-align: center; font-size: 20px'"></label>
      <span v-if="err" role="alert" style="color: #C0392B; font-size: 13px">{{ err }}</span>
      <button type="submit" :style="primary">Continue</button>
      <button type="button" :disabled="cool > 0" @click="resend" style="border: 0; background: transparent; color: #2F7A12; font-weight: 600; cursor: pointer; font-size: 14px">{{ cool > 0 ? 'Resend in ' + cool + 's' : 'Resend code' }}</button>
      <button type="button" @click="reset()" style="border: 0; background: transparent; color: var(--k-55645C,#55645C); cursor: pointer; font-size: 13px">Start over</button>
    </form>

    <form v-else-if="st.stage === 'quiz'" novalidate aria-label="Security questions" @submit.prevent="checkQuiz" style="display: flex; flex-direction: column; gap: 14px">
      <h1 :style="head">Security questions</h1>
      <label v-for="(q, i) in questions" :key="i" :style="label">{{ q }}<input v-model="answers[i]" type="text" autocomplete="off" :aria-label="q" :style="field"></label>
      <span v-if="err" role="alert" style="color: #C0392B; font-size: 13px">{{ err }}</span>
      <button type="submit" :disabled="busy" :style="primary">Continue</button>
    </form>

    <div v-else-if="st.stage === 'sent'" role="status" style="display: flex; flex-direction: column; gap: 14px">
      <h1 :style="head">Check your email</h1>
      <p style="margin: 0; color: var(--k-55645C,#55645C); font-size: 14px">If your answers were right, we sent a link and a code to {{ st.email }}. Open the link to choose a new password.</p>
      <button type="button" @click="reset()" style="border: 0; background: transparent; color: var(--k-55645C,#55645C); cursor: pointer; font-size: 13px">Start over</button>
    </div>

    <form v-else-if="st.stage === 'link'" novalidate aria-label="Choose a new password" @submit.prevent="setNew" style="display: flex; flex-direction: column; gap: 14px">
      <h1 :style="head">New password</h1>
      <label :style="label">Code<input v-model="code" type="text" inputmode="numeric" autocomplete="one-time-code" maxlength="6" placeholder="••••••" aria-label="Code" :style="field + '; letter-spacing: 0.4em; text-align: center; font-size: 20px'"></label>
      <label :style="label">New password<span style="position: relative; display: block"><input v-model="pw" :type="showPw ? 'text' : 'password'" autocomplete="new-password" aria-label="New password" :style="field + '; width: 100%; box-sizing: border-box; padding-right: 64px'"><button type="button" @click="showPw = !showPw" :aria-label="showPw ? 'Hide password' : 'Show password'" style="position: absolute; right: 8px; top: 8px; height: 28px; padding: 0 8px; border: 0; background: transparent; color: #2F7A12; font-weight: 600; cursor: pointer">{{ showPw ? 'Hide' : 'Show' }}</button></span></label>
      <label :style="label">Confirm password<input v-model="pw2" :type="showPw ? 'text' : 'password'" autocomplete="new-password" aria-label="Confirm password" :style="field"></label>
      <ul style="margin: 0; padding: 0; list-style: none; display: flex; flex-direction: column; gap: 2px; font-size: 12px"><li v-for="r in rules" :key="r[0]" :style="`color: ${r[1] ? '#2B7A12' : 'var(--k-8A988F,#8A988F)'}`">{{ r[1] ? '✓' : '○' }} {{ r[0] }}</li></ul>
      <span v-if="err" role="alert" style="color: #C0392B; font-size: 13px">{{ err }}</span>
      <button type="submit" :disabled="busy" :style="primary">Save new password</button>
    </form>

    <span style="text-align: center"><NuxtLink to="/login" style="color: #2F7A12; font-weight: 600">Sign in</NuxtLink></span>

  </div>
</template>
