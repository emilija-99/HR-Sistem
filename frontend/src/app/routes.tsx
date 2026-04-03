import type { ReactElement } from "react";
import HomePage from "../pages/Home/HomePage";

type AppRoute = {
  path: string;
  element: ReactElement;
};

export const routes: AppRoute[] = [{ path: "/", element: <HomePage /> }];
