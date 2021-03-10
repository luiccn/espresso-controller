import { makeStyles } from "@material-ui/core";
import AppBar from "@material-ui/core/AppBar";
import IconButton from "@material-ui/core/IconButton";
import Toolbar from "@material-ui/core/Toolbar";
import Typography from "@material-ui/core/Typography";
import GitHubIcon from "@material-ui/icons/GitHub";
import LocalCafeIcon from "@material-ui/icons/LocalCafe";
import React from "react";

const useStyles = makeStyles((theme) => ({
  toolbar: { paddingRight: theme.spacing(2) },
  menuButton: { marginRight: theme.spacing(2) },
  title: { flexGrow: 1 },
}));

export default function AppHeader() {
  const classes = useStyles();

  return (
    <AppBar position="absolute" >
      <Toolbar className={classes.toolbar}>
        <IconButton edge="start" color="inherit" className={classes.menuButton}>
          <LocalCafeIcon fontSize="large" />
        </IconButton>
        <Typography component="h1" variant="h6" color="inherit" noWrap className={classes.title}>
          Espresso Controller
        </Typography>
        <IconButton href="https://github.com/luiccn/espresso-controller" target="_blank" color="inherit">
          <GitHubIcon />
        </IconButton>
      </Toolbar>
    </AppBar>
  );
}
