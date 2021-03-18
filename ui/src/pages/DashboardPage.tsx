import Grid from "@material-ui/core/Grid";
import Paper from "@material-ui/core/Paper";
import Typography from "@material-ui/core/Typography";
import Chip from '@material-ui/core/Chip';
import Title from "../components/Title";
import { makeStyles, Box } from "@material-ui/core";
import moment from "moment";
import parsePromText, { Metric } from "parse-prometheus-text-format";
import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import ConfigurationDialog from "../components/ConfigurationDialog";
import MetricCard, { Severity } from "../components/MetricCard";
import TemperatureCard from "../components/TemperatureCard";
import TemperatureChart from "../components/TemperatureChart";
import { GetConfigurationRequest, TemperatureStreamRequest } from "../proto/pkg/espressopb/espresso_pb";
import { showConfigDialog } from "../redux/selectors";
import { endBoilerTemperatureStream, startBoilerTemperatureStream } from "../redux/boilerTemperatureSlice";
import { getConfiguration } from "../redux/configurationSlice";

const metricsRefreshIntervalMillis = 2000;

const useStyles = makeStyles((theme) => ({
  paper: {
    padding: theme.spacing(2),
    display: "flex",
    overflow: "auto",
    flexDirection: "column",
  },
  tallPaper: {
    padding: theme.spacing(2),
    display: "flex",
    overflow: "auto",
    flexDirection: "column",
    height: 300,
  },
  root: {
    display: 'flex',
    justifyContent: 'space-evenly',
    flexDirection: "row",
    flexWrap: 'wrap',
    '& > *': {
      margin: theme.spacing(0.5),
    }
  }
}));

export default () => {
  const showSetTemperatureModal = useSelector(showConfigDialog);

  const classes = useStyles();

  const d = useDispatch();

  const [open, setOpen] = React.useState(false);

  function handleOpen() {
    setOpen(true)
  }

  function handleClose() {
    setOpen(false)
  }

  //
  // Boiler temperature
  // ------------------
  useEffect(() => {
    d(startBoilerTemperatureStream(new TemperatureStreamRequest()));
    return () => {
      d(endBoilerTemperatureStream());
    };
  }, [d]);

  //
  // System metrics
  // --------------
  const [metricsRefreshedAt, setMetricsRefreshedAt] = useState<moment.Moment | undefined>();
  const [cpuTemperature, setCpuTemperature] = useState<number | undefined>();
  const [powerOn, setPowerOn] = useState<string | undefined>();
  const [powerStatus, setPowerStatus] = useState<PowerStatus | undefined>();

  interface PowerStatus {
    PowerSchedule: string;
    AutoOffDuration: number;
    OnSince: string;
    CurrentlyInASchedule: boolean;
    LastInteraction: string;
    PowerOn: boolean;
    StopScheduling: boolean;
  }

  const refreshMetrics = async () => {
    const metricsResp = await fetch("/metrics");
    const metricsRaw = await metricsResp.text();

    const powerResp = await fetch("/power/status");
    const power = await powerResp.json() as PowerStatus;

    const metricsMap: { [key: string]: Metric } = parsePromText(metricsRaw).reduce((acc, cur) => {
      return { ...acc, [cur.name]: cur };
    }, {});

    setMetricsRefreshedAt(moment());

    setCpuTemperature(parseFloat(metricsMap.espresso_raspi_cpu_temperature.metrics[0].value));

    setPowerOn(power.PowerOn ? "ON" : "OFF")
    setPowerStatus(power)
  };
  useEffect(() => {
    refreshMetrics();
    const interval = setInterval(refreshMetrics, metricsRefreshIntervalMillis);
    return () => {
      clearInterval(interval);
    };
  }, []);

  useEffect(() => {
    d(getConfiguration({ request: new GetConfigurationRequest() }));
  }, [d]);

  const getRaspiTemperatureSeverity = (raspiTemperature: number): Severity => {
    if (raspiTemperature < 80) {
      return "normal";
    } else if (80 <= raspiTemperature && raspiTemperature < 85) {
      return "warning";
    } else {
      return "error";
    }
  };

  function toggle() {
    setPowerOn("🤔")
    const requestOptions = { method: 'POST' };
    fetch("/power/toggle", requestOptions).catch(() => { });
  }

  return (
    <>
      {showSetTemperatureModal && <ConfigurationDialog />}
      <Grid container spacing={1}>
        <Grid item xs={12}>
          <Paper className={classes.tallPaper}>
            <TemperatureChart />
          </Paper>
        </Grid>

        {!open &&
          <Grid item xs={6}>
            <Paper className={classes.paper} onClick={handleOpen}>
              <MetricCard
                name="CPU 🌡️"
                value={cpuTemperature?.toFixed(2) ?? "--"}
                unitLabel="°C"
                asOf={metricsRefreshedAt}
                severity={cpuTemperature ? getRaspiTemperatureSeverity(cpuTemperature) : "normal"}
              />
            </Paper>
          </Grid>
        }
        {open &&
          <Grid item xs={12}>
            <Paper className={classes.paper} onClick={handleClose}>
              <Typography variant="h6">
                <Box color="primary.main">
                  <div className={classes.root}>
                    <Box m={2}>
                      <Chip variant="outlined" color="primary" label={"Schedule:  " + JSON.stringify(powerStatus?.PowerSchedule, null, "\t") ?? "--"} />
                    </Box>
                    <Box m={2}>
                      <Chip variant="outlined" color="primary" label={"Auto-off:  " + powerStatus?.AutoOffDuration ?? "--"} />
                    </Box>
                    <Box m={2}>
                      <Chip variant="outlined" color="primary" label={"Schedule on:  " + (powerStatus?.CurrentlyInASchedule ? "true" : "false") ?? "--"} />
                    </Box>
                    <Box m={2}>
                      <Chip variant="outlined" color="primary" label={"Stop scheduling:  " + (powerStatus?.StopScheduling ? "true" : "false") ?? "--"} />
                    </Box>
                    <Box m={2}>
                      <Chip variant="outlined" color="primary" label={"On Since  " + powerStatus?.OnSince ?? "--"} />
                    </Box>
                    <Box m={2}>
                      <Chip variant="outlined" color="primary" label={"Last interaction:  " + powerStatus?.LastInteraction ?? "--"} />
                    </Box>
                  </div>
                </Box>
              </Typography>
            </Paper>
          </Grid>
        }
        <Grid item xs={6}>
          <Paper className={classes.paper} onClick={toggle}>
            <Title>Power ⚡</Title>
            <Typography variant="h4">
              <Box color={powerOn === "ON" ? "success.main" : "warning.main"}>
                {powerOn ?? "--"}
              </Box>
            </Typography>
            {powerStatus?.OnSince !== "0 seconds" && (
              <Typography color="textSecondary">
                for {powerStatus?.OnSince.replace(" milliseconds", " ms").replace(" seconds", " s").replace(" minutes", "m").replace(" hours", "h")}
              </Typography>
            )}
            {powerStatus?.OnSince === "0 seconds" && metricsRefreshedAt && (
              <Typography color="textSecondary">
                as of {metricsRefreshedAt.format("HH:mm:ss")}
              </Typography>
            )}
          </Paper>
        </Grid>

        <TemperatureCard />

      </Grid>
    </>
  );
};
