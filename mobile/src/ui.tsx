import React from 'react';
import { StyleSheet, View } from 'react-native';

import { COLORS } from './config';

// the site's favicon: the lime rounded square with the dark H
export function Logo() {
  return (
    <View style={s.logo}>
      <View style={[s.bar, { left: 18.7, top: 16.5, width: 8.8, height: 37.4 }]} />
      <View style={[s.bar, { left: 42.9, top: 16.5, width: 8.8, height: 37.4 }]} />
      <View style={[s.bar, { left: 18.7, top: 30.8, width: 33, height: 8.8 }]} />
    </View>
  );
}

// "rgb(15, 23, 20)" -> "#0F1714" and whether it is dark
export function parseColor(c: string): { hex: string; dark: boolean } | null {
  const m = /rgba?\(\s*(\d+)[,\s]+(\d+)[,\s]+(\d+)/i.exec(c || '');
  if (!m) return null;
  const [r, g, b] = [m[1], m[2], m[3]].map((x) => Math.max(0, Math.min(255, Number(x))));
  const hex = '#' + [r, g, b].map((x) => x.toString(16).padStart(2, '0')).join('').toUpperCase();
  return { hex, dark: 0.2126 * r + 0.7152 * g + 0.0722 * b < 140 };
}

const s = StyleSheet.create({
  logo: { width: 70.4, height: 70.4, borderRadius: 16.5, backgroundColor: COLORS.accent, marginBottom: 8 },
  bar: { position: 'absolute', backgroundColor: COLORS.bg, borderRadius: 2.75 }
});
