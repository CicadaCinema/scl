<script lang="ts">
	import { onMount } from "svelte";

	import "carbon-components-svelte/css/all.css";
	import {
		Form,
		FormGroup,
		RadioButtonGroup,
		RadioButton,
		TextInput,
		Button,
		DataTable,
	} from "carbon-components-svelte";
	import type { DataTableRow } from "carbon-components-svelte/types/DataTable/DataTable.svelte";

	import Modal from "./components/Modal.svelte";

	// runs after the component is first rendered to the DOM
	onMount(async () => {
		console.log();

		const queryString = window.location.search;

		const urlParams = new URLSearchParams(queryString);

		// if we have returned from the strava oauth page
		// TODO: make authentication persistent!! - use local storage?!
		if (urlParams.has("code")) {
			const authorisation_url = "https://scl-api.vercel.app/api/authenticate";

			fetch(authorisation_url, {
				method: "POST",
				headers: {
					"Access-Control-Allow-Origin": "*",
					"Content-type": "application/json",
				},
				body: JSON.stringify({ client_code: urlParams.get("code") }),
			})
				.then((response) => response.json())
				.then((data) => {
					currentAccessToken = data["access_token"];
					loginModalOpen = false;
				})
				.catch((error) => {
					console.log(error);
					return [];
				});
		}
	});

	function stravaOauthRedirect() {
		// TODO: maybe not all these scopes are needed, but let's keep them here anyway for now
		const stravaLoginUrl = `https://www.strava.com/oauth/authorize?client_id=${process.env.STRAVA_CLIENT_ID}&response_type=code&redirect_uri=http://localhost:5000/&approval_prompt=force&scope=activity:read,activity:read_all,read,read_all`;
		window.location.href = stravaLoginUrl;
	}

	function submitForm() {
		invalidFields = [false, false, false];

		// parse first form field
		const re = /(?<=strava\.com\/clubs\/)(0|[1-9][0-9]*)/g;
		const reExec = re.exec(textinputvalue);

		if (reExec == null) {
			invalidFields[0] = true;
		}

		switch (activityLimitChoice) {
			case "activity-number":
				let parsedNum = parseInt(activityLimitValue[0]);
				if (isNaN(parsedNum)) {
					invalidFields[1] = true;
				} else if (!invalidFields[0]) {
					getClubData(
						currentAccessToken,
						reExec[0],
						false,
						parsedNum
					);
				}
				break;
			// TODO: specify that this can ONLY be YOUR OWN activity
			// I think the usefulness of this is questionable, but it allows easy record keeping if a
			// club admin, say, records a "base activity" to use as a reference point
			case "up-to-activity":
				const ree = /(?<=strava\.com\/activities\/)(0|[1-9][0-9]*)/g;
				console.log(ree);
				const reExecc = ree.exec(activityLimitValue[1]);
				console.log(activityLimitValue[1]);
				console.log(reExecc);
				if (reExecc == null) {
					invalidFields[2] = true;
				} else if (!invalidFields[0]) {
					getClubData(
						currentAccessToken,
						reExec[0],
						true,
						reExecc[0]
					);
				}

				break;
			default:
				console.error("Invalid activity limit selection!");
		}
	}

	function getClubData(
		user_token: string,
		club_id: string,
		activity_limit_method: boolean,
		activity_limit_parameter: any
	) {
		let request_url = `https://www.strava.com/api/v3/clubs/${club_id}/activities?page=1&per_page=200`;
		const auth_token = "Bearer " + user_token;
		// TODO: make these two arrays into one dictionary
		const known_field_names = [
			"name",
			"distance",
			"moving_time",
			"elapsed_time",
			"total_elevation_gain",
			"type",
			"workout_type",
		];
		let known_fields_values = [];

		// we better limit the number of activities
		if (!activity_limit_method) {
			request_url = `https://www.strava.com/api/v3/clubs/${club_id}/activities?page=1&per_page=${activity_limit_parameter}`;
		} else {
			// grab details of activity we are finishing with
			fetch(
				`https://www.strava.com/api/v3/activities/${activity_limit_parameter}`,
				{
					headers: {
						Authorization: auth_token,
					},
				}
			)
				.then((response) => response.json())
				.then((data) => {
					// record known values about chosen activity
					for (let known_field_name of known_field_names) {
						known_fields_values.push(data[known_field_name]);
					}
				})
				.catch((error) => {
					console.log(error);
					return [];
				});
		}

		fetch(request_url, {
			headers: {
				Authorization: auth_token,
			},
		})
			.then((response) => response.json())
			.then((data) => {
				// console.log(data);

				activityData = [];
				leaderboardData = [];
				activityData = activityData;
				leaderboardData = leaderboardData;

				let leaderboard = new Map<string, number>();

				let current_id = 0;

				for (let activity of data) {
					const thisDistance = activity["distance"];
					const thisAthlete = activity["athlete"]["firstname"];

					if (activity["type"] == "Run") {
						activityData.push({
							id: current_id.toString(),
							distance: (thisDistance / 1000).toFixed(2),
							athlete: thisAthlete,
						});

						if (leaderboard.has(thisAthlete)) {
							leaderboard.set(
								thisAthlete,
								leaderboard.get(thisAthlete) + thisDistance
							);
						} else {
							leaderboard.set(thisAthlete, thisDistance);
						}

						// if we are limiting based on an activity
						if (activity_limit_method) {
							let test_num = 0;
							for (var i = 0; i < known_field_names.length; i++) {
								if (
									activity[known_field_names[i]] !=
									known_fields_values[i]
								) {
									test_num += 1;
								}
							}
							if (test_num == 0) {
								// means we have hit the correct activity
								break;
							}
						}
					}

					current_id += 1;
				}

				current_id = 0;

				for (let [athlete, distance] of leaderboard) {
					leaderboardData.push({
						id: current_id.toString(),
						distance: (distance / 1000).toFixed(2),
						athlete: athlete,
					});
					current_id += 1;
				}
			})
			.catch((error) => {
				console.log(error);
				return [];
			});
	}

	let activityData: DataTableRow[] = [];
	let leaderboardData: DataTableRow[] = [];

	let loginModalOpen = true;

	// form related values
	let textinputvalue = "";
	let activityLimitChoice = "activity-number";
	let activityLimitValue = ["", ""];
	let invalidFields = [false, false, false];

	let currentAccessToken = "";

	// TODO: implement DataTable multi-select somehow, so you can quickly select a RANGE of activities
	// and then delete/include them as part of your leaderboard
</script>

<main>
	<h1>Strava Club Leaderboard</h1>

	<Form on:submit={submitForm} style="padding: 1em;">
		<FormGroup>
			<TextInput
				bind:invalid={invalidFields[0]}
				invalidText="Invalid value. Try copyting the club's URL."
				bind:value={textinputvalue}
				labelText="Club URL"
				placeholder="Enter club URL"
			/>
		</FormGroup>
		<FormGroup>
			<RadioButtonGroup
				legendText="Activity Limit"
				name="radio-button-group"
				bind:selected={activityLimitChoice}
			>
				<RadioButton
					id="radio-1"
					value="activity-number"
					labelText="Count latest activities by number"
				/>
				<RadioButton
					id="radio-2"
					value="up-to-activity"
					labelText="Count activities up to and including a specific one"
				/>
			</RadioButtonGroup>
		</FormGroup>
		<FormGroup>
			{#if activityLimitChoice == "activity-number"}
				<TextInput
					bind:invalid={invalidFields[1]}
					bind:value={activityLimitValue[0]}
					invalidText="Invalid value. Try entering a positive whole number."
					labelText="Activity number"
					placeholder="Enter number of recent activities to be counted"
				/>
			{/if}
			{#if activityLimitChoice == "up-to-activity"}
				<TextInput
					bind:invalid={invalidFields[2]}
					bind:value={activityLimitValue[1]}
					invalidText="Invalid value. Try copying an club member's activity URL."
					labelText="Earliest activity to count"
					placeholder="Enter an activity URL"
				/>
			{/if}
		</FormGroup>
		<Button type="submit">Submit</Button>
	</Form>

	<div class="my_row">
		<div class="my_column">
			<DataTable
				headers={[
					{ key: "distance", value: "Distance (km)" },
					{ key: "athlete", value: "Athlete" },
				]}
				rows={activityData}
				sortable={true}
			>
				<strong slot="title">Activities</strong></DataTable
			>
		</div>

		<div class="my_column">
			<DataTable
				headers={[
					{ key: "distance", value: "Distance (km)" },
					{ key: "athlete", value: "Athlete" },
				]}
				rows={leaderboardData}
				sortable={true}
			>
				<strong slot="title">Leaderboard</strong></DataTable
			>
		</div>
	</div>

	<Modal
		open={loginModalOpen}
		preventCloseOnClickOutside
		modalHeading="Strava authentication"
		primaryButtonText="Login"
		on:click:button--primary={stravaOauthRedirect}
		on:close
		on:open
		on:submit
	>
		<p>Please login with Strava.</p>
	</Modal>
</main>

<style>
	.my_row {
		display: flex;
	}

	.my_column {
		flex: 50%;
		padding: 1em;
	}

	h1 {
		padding: 1em;
		text-align: center;
		font-weight: bold;
	}
</style>
