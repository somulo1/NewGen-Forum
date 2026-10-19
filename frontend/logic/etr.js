// ETR (Kenya Electronic Tax Register) style tax invoice and receipt for Hub subscription invoices.
// The fiscal numbers (control unit serial, invoice number, signature) are generated here for the layout.
// A real fiscal device or eTIMS issues the legal ones; connect it and replace fiscalOf().
export const VAT_RATE = 0.16;
export const SELLER = { name: 'Hub Platforms Limited', pin: 'P052198341K', vat: '0123456A', address: 'Westlands Business Park, Nairobi', phone: '+254 20 000 0000', email: 'billing@hub.example', web: 'www.hub.example' };

const hash = (str) => { let h = 2166136261; for (let i = 0; i < str.length; i++) { h ^= str.charCodeAt(i); h = Math.imul(h, 16777619) >>> 0; } return h >>> 0; };
const hex = (n, len) => n.toString(16).toUpperCase().padStart(len, '0').slice(-len);

export function fiscalOf(no) {
  const h1 = hash(no), h2 = hash(no + 'cu'), h3 = hash(no + 'sig');
  const seq = String(h1 % 90000 + 10000);
  return { cuSerial: 'KRAMW' + String(h2).padStart(10, '0').slice(0, 10), cuInvoice: '00' + String(h1 % 900000 + 100000) + '/' + seq.padStart(8, '0'),
    signature: [hex(h1, 8), hex(h2, 8), hex(h3, 8), hex(h1 ^ h3, 8)].join('-'), verifyUrl: 'https://itax.kra.go.ke/KRA-Portal/invoiceChk.htm?actionCode=loadPage&invoiceNo=' + '00' + String(h1 % 900000 + 100000) + '/' + seq.padStart(8, '0') };
}

// Amounts are VAT inclusive. Returns lines with net, VAT and gross, plus totals.
export function splitVat(gross) { const net = Math.round(gross / (1 + VAT_RATE) * 100) / 100; return { net, vat: Math.round((gross - net) * 100) / 100, gross }; }

export function buildDoc({ invoice, kind, plan, seats, cycle, addons, buyer, method }) {
  const addGross = addons.map((a) => ({ desc: a.name + ' (monthly)', qty: 1, gross: a.price * 10 }));
  const addTotal = addGross.reduce((n, a) => n + a.gross, 0);
  const planGross = Math.max(0, invoice.amount - addTotal);
  const lines = [{ desc: 'Hub ' + plan.name + ' plan · ' + seats + ' seats (' + (cycle === 'annual' ? 'yearly' : 'monthly') + ')', qty: 1, gross: planGross }].concat(addGross).map((l) => ({ ...l, ...splitVat(l.gross), cls: 'A' }));
  const t = lines.reduce((m, l) => ({ net: m.net + l.net, vat: m.vat + l.vat, gross: m.gross + l.gross }), { net: 0, vat: 0, gross: 0 });
  const fiscal = fiscalOf(invoice.no);
  return { kind, title: kind === 'receipt' ? 'PAYMENT RECEIPT' : 'TAX INVOICE', invoiceNo: invoice.no, date: invoice.date, paid: invoice.status === 'Paid', status: invoice.status, seller: SELLER, buyer, lines, totals: { net: Math.round(t.net * 100) / 100, vat: Math.round(t.vat * 100) / 100, gross: Math.round(t.gross) }, method, fiscal, servedBy: 'Hub Billing' };
}
