
import './App.css'
import { Button } from '@/components/ui/button'
import { $blockLoadProfileStream, sendBlockLoadProfileRequest } from '@/store/dlms'
import { useStore } from '@nanostores/react';

function App() {

  const blockLoadProfileStream = useStore($blockLoadProfileStream);

  const handleClick = () => {
   sendBlockLoadProfileRequest()
  }

  return (
    <>
      <div className="flex min-h-svh flex-col items-center justify-center">
      <Button onClick={handleClick}>Click me</Button>

      {blockLoadProfileStream.map((profile) => (
        <div key={profile.meterIp}>
          <p>{profile.meterIp}</p>
          <p>{profile.profile.dateTime}</p>
          <p>{profile.profile.averageVoltage}</p>
          <p>{profile.profile.blockEnergyWhImport}</p>
          <p>{profile.profile.blockEnergyVahImport}</p>
          <p>{profile.profile.blockEnergyWhExport}</p>
          <p>{profile.profile.blockEnergyVahExport}</p>
          <p>{profile.profile.averageCurrent}</p>
          <p>{profile.profile.meterHealthIndicator}</p>
        </div>
      ))}
    </div>
    </>
  )
}

export default App
