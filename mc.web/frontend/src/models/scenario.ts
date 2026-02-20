export type ScenarioComponent = {
  assetId: number;
  weight: number;
};

export type Scenario = {
  id: number;
  name: string;
  floatedWeight: boolean;
  createdAt: string;
  updatedAt: string;
  components: ScenarioComponent[];
};

export type NewScenarioRequest = {
  name: string;
  floatedWeight: boolean;
  components: ScenarioComponent[];
};