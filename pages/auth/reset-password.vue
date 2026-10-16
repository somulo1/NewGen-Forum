<script setup>
// Reset password, step by step (no email server yet, so a "demo mailbox" stands in for the inbox):
// 1 email -> 2 code from email -> 3 security questions -> 4 link emailed -> 5 code from that email + new password -> sign in.
definePageMeta({ path: '/reset-password', layout: 'auth' });
useHead({ title: 'Reset password · Hub' });
const route = useRoute();
const strict = AUTH_STRICT;
const accounts = useAccounts();
const KEY = 'hub-reset';
const OTP_MS = 10 * 60 * 1000, LINK_MS = 30 * 60 * 1000;

const st = reactive({ stage: 'email', email: '', otp: '', otpAt: 0, otpTries: 0, quizTries: 0, token: '', otp2: '', otp2At: 0, otp2Tries: 0, mailbox: [], known: false });
const load = () => { try { Object.assign(st, JSON.parse(sessionStorage.getItem(KEY) || '{}')); } catch (e) { /* start fresh */ } };
const save = () => { try { sessionStorage.setItem(KEY, JSON.stringify(st)); } catch (e) { /* storage blocked */ } };
const reset = (msg = '') => { Object.assign(st, { stage: 'email', email: '', otp: '', otpAt: 0, otpTries: 0, quizTries: 0, token: '', otp2: '', otp2At: 0, otp2Tries: 0, mailbox: [], known: false }); notice.value = msg; err.value = ''; save(); };

const email = ref(''), code = ref(''), answers = ref(['', '', '']), pw = ref(''), pw2 = ref(''), showPw = ref(false);
const err = ref(''), notice = ref(''), busy = ref(false), cool = ref(0);
let timer = null;
const rnd = (n) => Array.from({ length: n }, () => Math.floor(Math.random() * 10)).join('');
const tok = () => Array.from(crypto.getRandomValues(new Uint8Array(16))).map((x) => x.toString(16).padStart(2, '0')).join('');
const startCool = () => { cool.value = 30; clearInterval(timer); timer = setInterval(() => { cool.value--; if (cool.value <= 0) clearInterval(timer); }, 1000); };
const mail = (m) => { st.mailbox.unshift({ ...m, at: new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' }) }); };

// Opened from the emailed link: it only works while the token matches and has not expired.
const openLink = () => {
  const t = route.query.token;
  if (!t) return;
  if (!strict || (st.token && t === st.token && Date.now() - st.otp2At < LINK_MS)) { st.stage = 'link'; code.value = ''; err.value = ''; save(); } else { reset('This link is no longer valid.'); }
};
onMounted(async () => { await accounts.seedDemo(); load(); openLink(); });
watch(() => route.query.token, () => openLink());
onBeforeUnmount(() => clearInterval(timer));

// 1. Email. The page says the same thing whether or not an account exists.
const sendCode = () => {
  err.value = ''; notice.value = '';
  const e = email.value.trim().toLowerCase();
  if (strict ? !/^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(e) : !e) { err.value = 'Enter your email.'; return; }
  const acc = accounts.get(e);
  Object.assign(st, { email: e, known: strict ? !!acc : true, otp: rnd(6), otpAt: Date.now(), otpTries: 0, quizTries: 0, stage: 'otp', mailbox: [] });
  if (st.known) mail({ to: e, subject: 'Your Hub code', body: '', code: st.otp });
  code.value = ''; startCool(); save();
};
const resend = () => { if (cool.value > 0) return; st.otp = rnd(6); st.otpAt = Date.now(); st.otpTries = 0; if (st.known) mail({ to: st.email, subject: 'Your Hub code', body: '', code: st.otp }); code.value = ''; err.value = ''; notice.value = 'Code sent.'; startCool(); save(); };

// 2. Code from the email.
const checkCode = () => {
  err.value = ''; notice.value = '';
  if (strict && Date.now() - st.otpAt > OTP_MS) { err.value = 'Incorrect code.'; return; }
  st.otpTries++;
  if (!strict || (st.known && code.value.trim() === st.otp)) { st.stage = 'quiz'; st.otpTries = 0; answers.value = ['', '', '']; save(); return; }
  if (st.otpTries >= 5) { reset('Try again.'); return; }
  err.value = 'Incorrect code.'; save();
};

// 3. Security questions. All answers must be right.
const acc = computed(() => accounts.get(st.email));
const questions = computed(() => (acc.value ? acc.value.quiz.map((x) => x.q) : SECURITY_QUESTIONS.slice(0, 3)));
const checkQuiz = async () => {
  err.value = ''; busy.value = true;
  if (strict && answers.value.some((a) => a.trim().length < 1)) { err.value = 'Answer all questions.'; busy.value = false; return; }
  const ok = strict ? await accounts.checkQuiz(st.email, answers.value) : true;
  busy.value = false;
  if (ok) {
    st.token = tok(); st.otp2 = rnd(6); st.otp2At = Date.now(); st.otp2Tries = 0; st.stage = 'sent';
    mail({ to: st.email, subject: 'Reset your Hub password', body: '', code: st.otp2, link: '/reset-password?token=' + st.token });
    save(); return;
  }
  st.quizTries++;
  if (st.quizTries >= 3) { reset('Try again.'); return; }
  err.value = 'Incorrect answers.'; save();
};

// 5. Code from the link email, then the new password.
const rules = computed(() => [['At least 8 characters', pw.value.length >= 8], ['An uppercase and a lowercase letter', /[a-z]/.test(pw.value) && /[A-Z]/.test(pw.value)], ['A number', /\d/.test(pw.value)]]);
const setNew = async () => {
  err.value = '';
  if (strict && Date.now() - st.otp2At > LINK_MS) { reset('This link is no longer valid.'); return; }
  st.otp2Tries++;
  if (strict && code.value.trim() !== st.otp2) { if (st.otp2Tries >= 5) { reset('Try again.'); return; } err.value = 'Incorrect code.'; save(); return; }
  if (strict && rules.value.some((r) => !r[1])) { err.value = 'Password does not meet the rules.'; return; }
  if (strict && pw.value !== pw2.value) { err.value = 'Passwords do not match.'; return; }
  busy.value = true; await accounts.setPassword(st.email, pw.value); busy.value = false;
  try { sessionStorage.removeItem(KEY); } catch (e) { /* nothing to clear */ }
  navigateTo('/login?reset=1');
};

const field = 'height: 44px; padding: 0 14px; border-radius: 12px; border: 1px solid #C9D6CF; background: #fff; color: #0F1714; font-size: 15px';
const primary = 'height: 46px; border-radius: 23px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 700; font-size: 15px; cursor: pointer';
const label = 'display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #55645C';
const head = "margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 26px; letter-spacing: -0.02em";
const steps = ['1', '2', '3', '4', '5'];
const stepNo = computed(() => ({ email: 0, otp: 1, quiz: 2, sent: 3, link: 4 })[st.stage]);
</script>
<template>
  <div style="display: flex; flex-direction: column; gap: 14px">
    <ol aria-label="Progress" style="list-style: none; margin: 0; padding: 0; display: flex; gap: 6px">
      <li v-for="(s, i) in steps" :key="s" :aria-current="i === stepNo ? 'step' : undefined" style="flex: 1; display: flex; flex-direction: column; gap: 4px; min-width: 0"><span :style="`height: 4px; border-radius: 2px; background: ${i <= stepNo ? '#7BC51F' : '#DCE5DF'}`"></span><span :style="`font-size: 10px; text-align: center; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: ${i === stepNo ? '#0F1714' : '#8A988F'}; font-weight: ${i === stepNo ? 700 : 500}`">{{ s }}</span></li>
    </ol>
    <div v-if="notice" role="status" style="padding: 10px 12px; border-radius: 10px; background: #EAF6DD; color: #2B5A0B; font-size: 13px">{{ notice }}</div>

    <form v-if="st.stage === 'email'" novalidate aria-label="Reset password" @submit.prevent="sendCode" style="display: flex; flex-direction: column; gap: 14px">
      <h1 :style="head">Reset password</h1>
      <label :style="label">Email<input v-model="email" type="email" autocomplete="email" placeholder="you@company.com" :style="field"></label>
      <span v-if="err" role="alert" style="color: #C0392B; font-size: 13px">{{ err }}</span>
      <button type="submit" :style="primary">Continue</button>
    </form>

    <form v-else-if="st.stage === 'otp'" novalidate aria-label="Enter the code" @submit.prevent="checkCode" style="display: flex; flex-direction: column; gap: 14px">
      <h1 :style="head">Enter code</h1>
      <label :style="label">Code<input v-model="code" type="text" inputmode="numeric" autocomplete="one-time-code" maxlength="6" placeholder="••••••" aria-label="Code" :style="field + '; letter-spacing: 0.4em; text-align: center; font-size: 20px'"></label>
      <span v-if="err" role="alert" style="color: #C0392B; font-size: 13px">{{ err }}</span>
      <button type="submit" :style="primary">Continue</button>
      <button type="button" :disabled="cool > 0" @click="resend" style="border: 0; background: transparent; color: #2F7A12; font-weight: 600; cursor: pointer; font-size: 14px">{{ cool > 0 ? 'Resend in ' + cool + 's' : 'Resend code' }}</button>
      <button type="button" @click="reset()" style="border: 0; background: transparent; color: #55645C; cursor: pointer; font-size: 13px">Start over</button>
    </form>

    <form v-else-if="st.stage === 'quiz'" novalidate aria-label="Security questions" @submit.prevent="checkQuiz" style="display: flex; flex-direction: column; gap: 14px">
      <h1 :style="head">Security questions</h1>
      <label v-for="(q, i) in questions" :key="i" :style="label">{{ q }}<input v-model="answers[i]" type="text" autocomplete="off" :aria-label="q" :style="field"></label>
      <span v-if="err" role="alert" style="color: #C0392B; font-size: 13px">{{ err }}</span>
      <button type="submit" :disabled="busy" :style="primary">Continue</button>
    </form>

    <div v-else-if="st.stage === 'sent'" role="status" style="display: flex; flex-direction: column; gap: 14px">
      <h1 :style="head">Check your email</h1>
      <button type="button" @click="reset()" style="border: 0; background: transparent; color: #55645C; cursor: pointer; font-size: 13px">Start over</button>
    </div>

    <form v-else-if="st.stage === 'link'" novalidate aria-label="Choose a new password" @submit.prevent="setNew" style="display: flex; flex-direction: column; gap: 14px">
      <h1 :style="head">New password</h1>
      <label :style="label">Code<input v-model="code" type="text" inputmode="numeric" autocomplete="one-time-code" maxlength="6" placeholder="••••••" aria-label="Code" :style="field + '; letter-spacing: 0.4em; text-align: center; font-size: 20px'"></label>
      <label :style="label">New password<span style="position: relative; display: block"><input v-model="pw" :type="showPw ? 'text' : 'password'" autocomplete="new-password" aria-label="New password" :style="field + '; width: 100%; box-sizing: border-box; padding-right: 64px'"><button type="button" @click="showPw = !showPw" :aria-label="showPw ? 'Hide password' : 'Show password'" style="position: absolute; right: 8px; top: 8px; height: 28px; padding: 0 8px; border: 0; background: transparent; color: #2F7A12; font-weight: 600; cursor: pointer">{{ showPw ? 'Hide' : 'Show' }}</button></span></label>
      <label :style="label">Confirm password<input v-model="pw2" :type="showPw ? 'text' : 'password'" autocomplete="new-password" aria-label="Confirm password" :style="field"></label>
      <ul style="margin: 0; padding: 0; list-style: none; display: flex; flex-direction: column; gap: 2px; font-size: 12px"><li v-for="r in rules" :key="r[0]" :style="`color: ${r[1] ? '#2B7A12' : '#8A988F'}`">{{ r[1] ? '✓' : '○' }} {{ r[0] }}</li></ul>
      <span v-if="err" role="alert" style="color: #C0392B; font-size: 13px">{{ err }}</span>
      <button type="submit" :disabled="busy" :style="primary">Save new password</button>
    </form>

    <span style="text-align: center"><NuxtLink to="/login" style="color: #2F7A12; font-weight: 600">Sign in</NuxtLink></span>

    <details style="border: 1px dashed #B7C6BD; border-radius: 12px; padding: 10px 12px; background: #FAFCFB" :open="st.mailbox.length > 0">
      <summary style="cursor: pointer; font-size: 12px; color: #8A988F">Demo mailbox</summary>
      <div style="display: flex; flex-direction: column; gap: 8px; margin-top: 10px">
        <div v-for="(m, i) in st.mailbox" :key="i" style="padding: 10px 12px; border-radius: 10px; background: #fff; border: 1px solid #DCE5DF; display: flex; flex-direction: column; gap: 4px; font-size: 12px">
          <strong style="font-size: 13px; color: #0F1714">{{ m.subject }}</strong>
          <span v-if="m.code" style="font-family: ui-monospace, monospace; font-size: 18px; letter-spacing: 0.3em; color: #0F1714">{{ m.code }}</span>
          <NuxtLink v-if="m.link" :to="m.link" style="color: #2F7A12; font-weight: 600">Reset password</NuxtLink>
        </div>
      </div>
    </details>
  </div>
</template>
