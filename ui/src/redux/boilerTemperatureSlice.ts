import { createSlice, Dispatch, PayloadAction } from "@reduxjs/toolkit";
import moment from "moment";
import { toast } from "react-toastify";
import { RootState } from ".";
import { TemperatureStreamRequest, TemperatureStreamResponse } from "../proto/pkg/espressopb/espresso_pb";
import { EspressoClient } from "../proto/pkg/espressopb/espresso_pb_service";
import { TemperatureSample } from "../types";

type ReturnType<T extends (...args: any[]) => any> = T extends (...args: any[]) => infer R ? R : never;

const espressoClient = new EspressoClient("");

const maxNumSamples = 900; // 15 minutes of history at a rate of 1 sample/sec

//
// SLICE
// -----

interface BoilerTemperatureSlice {
  stream?: ReturnType<typeof espressoClient.boilerTemperature>;
  history: TemperatureSample[];
}

const initialState: BoilerTemperatureSlice = {
  stream: undefined,
  history: [],
};

const boilerTemperatureSlice = createSlice({
  name: "boilerTemperature",
  initialState,
  reducers: {
    // BoilerTemperature server streaming rpc
    startBoilerTemperatureStream: (
      state,
      action: PayloadAction<ReturnType<typeof espressoClient.boilerTemperature>>
    ) => {
      state.stream = action.payload;
    },
    receiveBoilerTemperatureStreamMsg: (state, action: PayloadAction<TemperatureStreamResponse>) => {
      const msg = action.payload;
      switch (msg.getDataCase()) {
        case TemperatureStreamResponse.DataCase.HISTORY:
          const history = msg.getHistory();
          if (!history) {
            throw new Error("Failed to process boiler temperature stream message data");
          }

          state.history = history
            .getSamplesList()
            .reduce((acc: TemperatureSample[], curSample) => {
              const observedAt = curSample.getObservedAt();
              return observedAt
                ? [
                    ...acc,
                    {
                      value: curSample.getValue(),
                      observedAt: moment(observedAt.toDate()),
                    },
                  ]
                : acc;
            }, [])
            .filter((s) => s !== null);
          break;
        case TemperatureStreamResponse.DataCase.SAMPLE:
          const sample = msg.getSample();
          if (!sample) {
            throw new Error("Failed to process boiler temperature stream message data");
          }

          const observedAt = sample.getObservedAt();
          if (!observedAt) {
            throw new Error("Failed to process boiler temperature stream message data");
          }

          state.history = state.history.slice(-(maxNumSamples - 1));
          state.history.push({
            value: sample.getValue(),
            observedAt: moment(observedAt.toDate()),
          });
          break;
        default:
          throw new Error("Failed to process boiler temperature stream message data");
      }
    },
    endBoilerTemperatureStream: (state) => {
      state.stream?.cancel();
      state.stream = undefined;
    },
  },
});

export default boilerTemperatureSlice.reducer;

export const startBoilerTemperatureStream = (req: TemperatureStreamRequest) => (d: Dispatch) => {
  const stream = espressoClient.boilerTemperature(req);
  d(boilerTemperatureSlice.actions.startBoilerTemperatureStream(stream));

  stream.on("data", (msg) => {
    try {
      d(boilerTemperatureSlice.actions.receiveBoilerTemperatureStreamMsg(msg));
    } catch (e) {
      toast.error(`Error: ${e.message}`);
    }
  });
};

export const { endBoilerTemperatureStream } = boilerTemperatureSlice.actions;

//
// SELECTORS
// ---------

export const selectCurTemp = (state: RootState) =>
  state.boilerTemperature.history[state.boilerTemperature.history.length - 1];
export const selectTempHistory = (state: RootState) => state.boilerTemperature.history;
