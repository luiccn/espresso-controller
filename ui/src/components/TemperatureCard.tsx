import { Button, Grid, Paper } from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
import React from "react";
import { useDispatch, useSelector } from "react-redux";
import { selectCurTemp } from "../redux/boilerTemperatureSlice";
import { selectConfiguration } from "../redux/configurationSlice";
import { setConfigureDialogVisibility } from "../redux/uiSlice";
import Title from "./Title";

const useStyles = makeStyles({
  temperatureContext: {
    flex: 1,
  },
  setTargetTempButton: { textAlign: "center" },
});

export default function TemperatureCard() {
  const d = useDispatch();
  const classes = useStyles();

  const curTemp = useSelector(selectCurTemp);
  const configuration = useSelector(selectConfiguration);

  const handleConfigureClicked = () => d(setConfigureDialogVisibility(true));

  return (
    <>

      <Grid container spacing={1} >
        <Grid item xs={6}>
          <Paper>
            <Title>Boiler 🌡️</Title>
            <Typography variant="h4" color="primary">
              {curTemp?.value.toFixed(2) ?? "--"} °C
        </Typography>
            {curTemp && (
              <Typography color="textSecondary" className={classes.temperatureContext}>
                as of {curTemp.observedAt.format("HH:mm:ss")}
              </Typography>
            )}
          </Paper>
        </Grid>

        <Grid item xs={6}>
          <Paper>
            <Title>Target 🌡️</Title>
            <Typography variant="h4" color="primary">
              {configuration?.targetTemp.value} °C
      </Typography>
            <Typography color="textSecondary" className={classes.temperatureContext}>
              set {configuration?.targetTemp.setAt.fromNow()}
            </Typography>
          </Paper>
        </Grid>
        <Grid item xs={12}>
          <Paper className={classes.setTargetTempButton}>
            <Button variant="contained" color="primary" size="small" onClick={handleConfigureClicked}>
              CONFIGURE PID
          </Button>
          </Paper>
        </Grid>
      </Grid>
    </>
  );
}
