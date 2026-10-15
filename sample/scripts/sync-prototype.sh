#!/usr/bin/env bash
# Re-port the prototype after it changes in Claude Design.
#   1) Put the new Main.dc.html and Table.dc.html into prototype-source/
#   2) npm run sync
# WARNING: this regenerates layouts/, pages/, components/{layout,sections,panels,overlays}/ and logic/.
# Commit your work first; re-apply hand edits with git diff afterwards.
set -e
node scripts/convert-dc.cjs prototype-source/Main.dc.html prototype-source/generated/HubApp.vue HubApp
node scripts/convert-dc.cjs prototype-source/Table.dc.html components/ui/DataTable.vue DataTable
mv prototype-source/generated/HubApp.css assets/hub.css
rm -f components/ui/DataTable.css
node scripts/split-prototype.cjs
echo "Done. Run npm run dev and compare with design-reference/."
