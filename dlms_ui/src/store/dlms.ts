import { atom, onMount } from "nanostores";
import { DLMSProcessorClient,GetBlockLoadProfileRequest, GetBlockLoadProfileResponse, Meter } from "../proto/dlmsprocessor";

// Create chat client with JWT authentication
var dlmsProcessor = new DLMSProcessorClient(
  import.meta.env.VITE_API_URL,
  {},
);

// Meter configuration store
export const $meterConfig = atom<Meter>(Meter.fromObject({
  ip: "2401:4900:833f:2688:0000:0000:0000:0002",
  port: 4059,
  authPassword: "0000000000000000",
  systemTitle: "6162636465666768",
  blockCipherKey: "49423031494230324942303349423034",
  authKey: "49423031494230324942303349423034",
  clientAddress: "48",
  serverAddress: "1",
  obis: "1.0.1.8.0.255",
}));

export const $blockLoadProfileStream = atom<GetBlockLoadProfileResponse[]>([]);

export const sendBlockLoadProfileRequest = () => {
    $blockLoadProfileStream.set([]);
  const blockLoadProfileStream = dlmsProcessor.GetBlockLoadProfile(
    GetBlockLoadProfileRequest.fromObject({    
        meter: [
            $meterConfig.get()
        ],
        retries: 3,
        connectionTimeout: 60,
        retryDelay: 5,
    })
    , {}
)

blockLoadProfileStream.on("data", (profile: GetBlockLoadProfileResponse) => {
    console.log(profile);
    $blockLoadProfileStream.set([...$blockLoadProfileStream.get(), profile]);
})

}