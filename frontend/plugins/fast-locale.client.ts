// Date.prototype.toLocale*String builds a new Intl.DateTimeFormat on every call, which is slow (the Security log and ticket
// tables format hundreds of dates on every redraw, and typing redrew them all). This keeps one formatter per locale and
// option set and gives the same text, so the app stays responsive while someone types.
const cache = new Map<string, Intl.DateTimeFormat>();
type Kind = 'date' | 'time' | 'any';
const has = (o: any, keys: string[]) => keys.some((k) => o[k] !== undefined);
const DATE = ['weekday', 'year', 'month', 'day'];
const TIME = ['dayPeriod', 'hour', 'minute', 'second', 'fractionalSecondDigits'];

function formatter(kind: Kind, locale: any, options: any): Intl.DateTimeFormat | null {
  const o: any = options ? { ...options } : {};
  if (o.dateStyle !== undefined || o.timeStyle !== undefined) return null;   // styles have their own rules: use the built-in path
  const noDate = !has(o, DATE), noTime = !has(o, TIME);
  if ((kind === 'date' && noDate) || (kind === 'any' && noDate && noTime)) Object.assign(o, { year: 'numeric', month: 'numeric', day: 'numeric' });
  if ((kind === 'time' && noTime) || (kind === 'any' && noDate && noTime)) Object.assign(o, { hour: 'numeric', minute: 'numeric', second: 'numeric' });
  const key = kind + '|' + (Array.isArray(locale) ? locale.join(',') : String(locale)) + '|' + JSON.stringify(o);
  let f = cache.get(key);
  if (!f) {
    try { f = new Intl.DateTimeFormat(locale, o); } catch (e) { return null; }
    cache.set(key, f);
  }
  return f;
}

export default defineNuxtPlugin(() => {
  const patch = (name: 'toLocaleDateString' | 'toLocaleTimeString' | 'toLocaleString', kind: Kind) => {
    const original = Date.prototype[name];
    (Date.prototype as any)[name] = function (this: Date, locale?: any, options?: any) {
      if (isNaN(this.getTime())) return 'Invalid Date';
      const f = formatter(kind, locale, options);
      return f ? f.format(this) : (original as any).call(this, locale, options);
    };
  };
  patch('toLocaleDateString', 'date');
  patch('toLocaleTimeString', 'time');
  patch('toLocaleString', 'any');
});
