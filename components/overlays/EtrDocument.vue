<!-- ETR style tax invoice (A4) and receipt (80 mm roll) for subscription invoices. Print or save as PDF from the browser. -->
<template><template v-if="d.open">
<div class="hub-overlay" role="dialog" aria-modal="true" aria-label="Print invoice" style="position: fixed; inset: 0; z-index: 75; background: rgba(5,10,8,0.82); display: flex; flex-direction: column">
<div style="flex: none; display: flex; flex-wrap: wrap; align-items: center; gap: 8px 12px; padding: 12px 16px; color: #E8EFEA">
<strong style="flex: 1 1 160px; min-width: 0; font-size: 15px">{{ d.invoiceNo }}</strong>
<div role="tablist" aria-label="Document type" style="display: flex; gap: 4px"><button role="tab" :aria-selected="d.kind === 'invoice'" @click="d.switchKind('invoice')" :style="tab(d.kind === 'invoice')">Tax invoice</button><button v-if="d.paid" role="tab" :aria-selected="d.kind === 'receipt'" @click="d.switchKind('receipt')" :style="tab(d.kind === 'receipt')">Receipt</button></div>
<button @click="print" style="height: 38px; padding: 0 20px; border-radius: 19px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Print / Save as PDF</button>
<button aria-label="Close" @click="d.close" style="width: 38px; height: 38px; border-radius: 50%; border: 1px solid #3A4A42; background: transparent; color: #E8EFEA; cursor: pointer">✕</button>
</div>
<p style="flex: none; margin: 0; padding: 0 16px 8px; font-size: 12px; color: #9DB0A6">Laid out like a KRA ETR document. The control unit numbers and signature are placeholders until an ETR or eTIMS device is connected to issue the real ones.</p>
<div style="flex: 1; min-height: 0; overflow: auto; padding: 4px 16px 24px; display: flex; justify-content: center; align-items: flex-start">
<div id="etr-paper" ref="paper" :class="d.kind === 'receipt' ? 'etr etr-receipt' : 'etr etr-invoice'">

<!-- ===== RECEIPT (80 mm thermal roll) ===== -->
<template v-if="d.kind === 'receipt'">
<div class="etr-c etr-b etr-lg">{{ d.seller.name }}</div>
<div class="etr-c">{{ d.seller.address }}</div>
<div class="etr-c">Tel: {{ d.seller.phone }}</div>
<div class="etr-c">PIN: {{ d.seller.pin }} &nbsp; VAT: {{ d.seller.vat }}</div>
<div class="etr-dash"></div>
<div class="etr-c etr-b etr-md">{{ d.title }}</div>
<div class="etr-row"><span>Receipt for</span><span>{{ d.invoiceNo }}</span></div>
<div class="etr-row"><span>Date</span><span>{{ dateLabel }}</span></div>
<div class="etr-row"><span>Served by</span><span>{{ d.servedBy }}</span></div>
<div class="etr-row"><span>Customer</span><span class="etr-r">{{ d.buyer.name }}</span></div>
<div class="etr-row"><span>Customer PIN</span><span>{{ d.buyer.pin }}</span></div>
<div class="etr-dash"></div>
<template v-for="(l, i) in d.lines" :key="i"><div>{{ l.desc }}</div><div class="etr-row"><span>{{ l.qty }} x {{ money(l.gross) }}</span><span>{{ money(l.gross) }} {{ l.cls }}</span></div></template>
<div class="etr-dash"></div>
<div class="etr-row etr-b"><span>TOTAL (KES)</span><span>{{ money(d.totals.gross) }}</span></div>
<div class="etr-row"><span>{{ d.method }}</span><span>{{ money(d.totals.gross) }}</span></div>
<div class="etr-row"><span>Change</span><span>0.00</span></div>
<div class="etr-dash"></div>
<div class="etr-row etr-b"><span>Tax</span><span>Taxable</span><span>VAT</span></div>
<div class="etr-row"><span>A 16%</span><span>{{ money(d.totals.net) }}</span><span>{{ money(d.totals.vat) }}</span></div>
<div class="etr-dash"></div>
<div class="etr-c etr-sm">Prices are VAT inclusive</div>
<div class="etr-c etr-sm">Items marked A are taxed at 16%</div>
<div class="etr-dash"></div>
<div class="etr-c etr-b">FISCAL DETAILS</div>
<div class="etr-kv"><span>CU Serial No</span><span>{{ d.fiscal.cuSerial }}</span></div>
<div class="etr-kv"><span>CU Invoice No</span><span>{{ d.fiscal.cuInvoice }}</span></div>
<div class="etr-kv"><span>Signature</span><span class="etr-break">{{ d.fiscal.signature }}</span></div>
<div class="etr-qr" v-html="qr"></div>
<div class="etr-c etr-sm">Scan to verify on the KRA iTax portal</div>
<div class="etr-c etr-b" style="margin-top: 6px">Thank you for your payment</div>
<div class="etr-c etr-sm">{{ d.seller.email }}</div>
</template>

<!-- ===== TAX INVOICE (A4) ===== -->
<template v-else>
<div class="etr-head">
<div><div class="etr-xl etr-b">{{ d.seller.name }}</div><div>{{ d.seller.address }}</div><div>Tel: {{ d.seller.phone }} · {{ d.seller.email }}</div><div>PIN: <b>{{ d.seller.pin }}</b> &nbsp; VAT No: <b>{{ d.seller.vat }}</b></div></div>
<div class="etr-r"><div class="etr-xl etr-b">{{ d.title }}</div><div>No. <b>{{ d.invoiceNo }}</b></div><div>Date: {{ dateLabel }}</div><div><span :class="['etr-badge', d.paid ? 'etr-paid' : 'etr-due']">{{ d.status.toUpperCase() }}</span></div></div>
</div>
<div class="etr-parties">
<div class="etr-box"><div class="etr-cap">Bill to</div><div class="etr-b">{{ d.buyer.name }}</div><div>{{ d.buyer.address }}</div><div>{{ d.buyer.email }}</div><div>Customer PIN: <b>{{ d.buyer.pin }}</b></div></div>
<div class="etr-box"><div class="etr-cap">Payment</div><div>Method: {{ d.method }}</div><div>Currency: Kenya Shillings (KES)</div><div>Terms: Due on receipt</div></div>
</div>
<table class="etr-table"><thead><tr><th>Description</th><th class="etr-r">Qty</th><th class="etr-r">Unit price</th><th class="etr-r">Tax</th><th class="etr-r">Amount</th></tr></thead>
<tbody><template v-for="(l, i) in d.lines" :key="i"><tr><td>{{ l.desc }}</td><td class="etr-r">{{ l.qty }}</td><td class="etr-r">{{ money(l.gross) }}</td><td class="etr-r">{{ l.cls }}</td><td class="etr-r">{{ money(l.gross) }}</td></tr></template></tbody></table>
<div class="etr-bottom">
<div class="etr-fiscal"><div class="etr-qr" v-html="qr"></div>
<div class="etr-fiscal-text"><div class="etr-cap">Fiscal details</div><div class="etr-kv"><span>CU Serial No</span><span>{{ d.fiscal.cuSerial }}</span></div><div class="etr-kv"><span>CU Invoice No</span><span>{{ d.fiscal.cuInvoice }}</span></div><div class="etr-kv"><span>Signature</span><span class="etr-break">{{ d.fiscal.signature }}</span></div><div class="etr-sm">Scan the code to verify this invoice on the KRA iTax portal.</div></div></div>
<div class="etr-totals"><div class="etr-row"><span>Taxable amount (A 16%)</span><span>{{ money(d.totals.net) }}</span></div><div class="etr-row"><span>VAT 16%</span><span>{{ money(d.totals.vat) }}</span></div><div class="etr-row etr-b etr-total"><span>TOTAL (KES)</span><span>{{ money(d.totals.gross) }}</span></div><div class="etr-sm">Prices are VAT inclusive. Tax class A = 16%.</div></div>
</div>
<div class="etr-foot">Served by {{ d.servedBy }} · {{ d.seller.web }} · This is a computer generated tax invoice.</div>
</template>
</div>
</div>
</div>
</template></template>

<script>
import QRCode from 'qrcode';
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'EtrDocument',
  data() { return { qr: '' }; },
  computed: { ...hubBindings(['settingsPage']).computed, d() { return this.settingsPage.etr; }, dateLabel() { return this.d.dateLabel; } },
  watch: { 'd.fiscal.verifyUrl': { immediate: true, handler(u) { this.makeQr(u); } } },
  methods: {
    money(n) { return Number(n).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }); },
    tab(on) { return `height: 36px; padding: 0 16px; border-radius: 18px; cursor: pointer; font-size: 13px; border: 1px solid ${on ? '#B8F36B' : '#3A4A42'}; background: ${on ? '#2C3D34' : 'transparent'}; color: #E8EFEA; font-weight: ${on ? 600 : 500}`; },
    async makeQr(u) { if (!u) { this.qr = ''; return; } try { this.qr = await QRCode.toString(u, { type: 'svg', margin: 0, width: 120, errorCorrectionLevel: 'M' }); } catch (e) { this.qr = ''; } },
    // Prints only the paper: a copy goes into its own element and everything else is hidden while printing.
    print() {
      const root = document.createElement('div'); root.id = 'etr-print-root'; root.innerHTML = this.$refs.paper.outerHTML;
      const st = document.createElement('style'); st.id = 'etr-print-style';
      st.textContent = '@page{size:' + (this.d.kind === 'receipt' ? '80mm auto;margin:3mm' : 'A4;margin:12mm') + '}@media print{body > *:not(#etr-print-root){display:none !important}#etr-print-root{display:block !important}}@media screen{#etr-print-root{display:none}}';
      document.body.appendChild(st); document.body.appendChild(root);
      const clean = () => { root.remove(); st.remove(); window.removeEventListener('afterprint', clean); };
      window.addEventListener('afterprint', clean); window.print(); setTimeout(clean, 4000);
    }
  }
};
</script>

<style>
.etr { background: #fff; color: #111; font-family: 'Courier New', ui-monospace, monospace; font-size: 12px; line-height: 1.45; box-sizing: border-box; }
.etr-receipt { width: 302px; max-width: 100%; padding: 14px 12px 18px; box-shadow: 0 8px 30px rgba(0,0,0,.5); }
.etr-invoice { width: 794px; max-width: 100%; padding: 36px 40px; font-family: Arial, Helvetica, sans-serif; font-size: 13px; box-shadow: 0 8px 30px rgba(0,0,0,.5); }
.etr b, .etr .etr-b { font-weight: 700; }
.etr .etr-c { text-align: center; }
.etr .etr-r { text-align: right; }
.etr .etr-sm { font-size: 10.5px; color: #444; }
.etr .etr-md { font-size: 14px; letter-spacing: .06em; margin: 2px 0 6px; }
.etr .etr-lg { font-size: 15px; }
.etr .etr-xl { font-size: 20px; }
.etr .etr-dash { border-top: 1px dashed #111; margin: 6px 0; }
.etr .etr-row { display: flex; justify-content: space-between; gap: 8px; }
.etr .etr-row > span:last-child { text-align: right; }
.etr .etr-kv { display: flex; justify-content: space-between; gap: 8px; font-size: 11px; }
.etr .etr-kv > span:last-child { text-align: right; }
.etr .etr-break { word-break: break-all; }
.etr .etr-qr { display: flex; justify-content: center; margin: 8px 0 4px; }
.etr .etr-qr svg { width: 120px; height: 120px; }
.etr-invoice .etr-head { display: flex; justify-content: space-between; gap: 24px; padding-bottom: 14px; border-bottom: 2px solid #111; }
.etr-invoice .etr-badge { display: inline-block; margin-top: 6px; padding: 2px 10px; border: 2px solid; font-weight: 700; font-size: 12px; letter-spacing: .08em; }
.etr-invoice .etr-paid { color: #0a7a3d; }
.etr-invoice .etr-due { color: #b45309; }
.etr-invoice .etr-parties { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin: 18px 0; }
.etr-invoice .etr-box { border: 1px solid #bbb; padding: 10px 12px; }
.etr-invoice .etr-cap { font-size: 10.5px; letter-spacing: .08em; text-transform: uppercase; color: #555; margin-bottom: 4px; }
.etr-invoice .etr-table { width: 100%; border-collapse: collapse; margin-bottom: 18px; }
.etr-invoice .etr-table th { background: #eee; text-align: left; padding: 8px 10px; border-bottom: 1px solid #999; }
.etr-invoice .etr-table th.etr-r { text-align: right; }
.etr-invoice .etr-table td { padding: 8px 10px; border-bottom: 1px solid #ddd; }
.etr-invoice .etr-bottom { display: flex; justify-content: space-between; gap: 24px; align-items: flex-start; }
.etr-invoice .etr-fiscal { display: flex; gap: 14px; align-items: flex-start; flex: 1 1 280px; min-width: 0; }
.etr-invoice .etr-fiscal .etr-qr { margin: 0; flex: none; }
.etr-invoice .etr-fiscal-text { min-width: 0; flex: 1; font-size: 11px; }
.etr-invoice .etr-totals { width: 260px; flex: none; display: flex; flex-direction: column; gap: 4px; }
.etr-invoice .etr-total { border-top: 2px solid #111; padding-top: 6px; font-size: 15px; }
.etr-invoice .etr-foot { margin-top: 24px; padding-top: 10px; border-top: 1px solid #bbb; text-align: center; font-size: 11px; color: #555; }
@media (max-width: 700px) { .etr-invoice { padding: 20px 16px; } .etr-invoice .etr-head, .etr-invoice .etr-bottom { flex-direction: column; } .etr-invoice .etr-parties { grid-template-columns: 1fr; } .etr-invoice .etr-totals { width: 100%; } .etr-invoice .etr-r { text-align: left; } .etr-invoice .etr-table .etr-r { text-align: right; } }
</style>
