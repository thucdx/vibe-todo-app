import PinGate from './components/auth/PinGate'
import AppShell from './components/layout/AppShell'

export default function App() {
  return (
    <PinGate>
      <AppShell />
    </PinGate>
  )
}
