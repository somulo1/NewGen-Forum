import { SafeAreaProvider } from 'react-native-safe-area-context';

import HubShell from './src/HubShell';

export default function App() {
  return (
    <SafeAreaProvider>
      <HubShell />
    </SafeAreaProvider>
  );
}
