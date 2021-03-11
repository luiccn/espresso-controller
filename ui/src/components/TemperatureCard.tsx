import { Grid, Paper } from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
import React from "react";
import { useDispatch, useSelector } from "react-redux";
import { selectCurTemp } from "../redux/boilerTemperatureSlice";
import { selectConfiguration } from "../redux/configurationSlice";
import { setConfigureDialogVisibility } from "../redux/uiSlice";
import Title from "./Title";

const useStyles = makeStyles((theme) => ({
  paper: {
    padding: theme.spacing(2),
    display: "flex",
    overflow: "auto",
    flexDirection: "column",
  },
  temperatureContext: {
    flex: 1,
  },
  setTargetTempButton: { textAlign: "center" },
}));

export default function TemperatureCard() {
  const d = useDispatch();
  const classes = useStyles();

  const curTemp = useSelector(selectCurTemp);
  const configuration = useSelector(selectConfiguration);

  const handleConfigureClicked = () => d(setConfigureDialogVisibility(true));

  return (
    <>

        <Grid item xs={6}>
          <Paper className={classes.paper}>
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
          <Paper className={classes.paper} onClick={handleConfigureClicked}>
            <Title>Target 🌡️</Title>
            <Typography variant="h4" color="primary">
              {configuration?.targetTemp.value} °C
      </Typography>
            <Typography color="textSecondary" className={classes.temperatureContext}>
              set {configuration?.targetTemp.setAt.fromNow()}
            </Typography>
          </Paper>
        </Grid>
    </>
  );
}
