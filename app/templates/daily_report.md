### Date Range: [StartDate] - [EndDate]

## Activity

- Total Active Time: [LLM to fill from activities_summaries.active_minutes]
- Calories Burned: [LLM to fill from activities_summaries.calories_out]
- Steps Taken: [LLM to fill from activities_summaries.steps]
- Distance Traveled: [LLM to fill from activities_summaries.distance / activities_summary_distances.distance]


### Active Minutes Breakdown

- Lightly Active Minutes: [LLM to fill from activities_summaries.lightly_active_minutes]
- Fairly Active Minutes: [LLM to fill from activities_summaries.fairly_active_minutes]
- Very Active Minutes: [LLM to fill from activities_summaries.very_active_minutes]

### Heart Rate Zones

- [LLM to iterate through activities_summary_heart_rate_zones and fill from heart_rate_zones (zone name, minutes)]

## Sleep

- Total Sleep Duration: [LLM to fill from sleep_logs.duration]
- Sleep Quality: [LLM to fill from sleep_logs.efficiency]
- Deep Sleep: [LLM to fill from sleep_stage_details where sleep_stage='deep sleep'] (minutes)
- Light Sleep: [LLM to fill from sleep_stage_details where sleep_stage='light sleep'] (minutes)
- REM Sleep: [LLM to fill from sleep_stage_details where sleep_stage='rem sleep'] (minutes)
- Time to Fall Asleep: [LLM to fill from sleep_logs.minutes_to_fall_asleep]


## Exercise Activities

- [LLM to iterate through daily_activity_summary_activities / minimal_activities and fill name, duration, calories burned (from activity_logs)]

##  Goals

### Activity Goals

- Calories Out Goal: [LLM to fill from goals.calories_out]
- Distance Goal: [LLM to fill from goals.distance]
- Steps Goal: [LLM to fill from goals.steps]

### Progress Towards Goals:

- Calories Out: [LLM to calculate based on activities_summaries.calories_out and goals.calories_out]
- Distance: [LLM to calculate based on activities_summaries.distance and goals.distance]
- Steps: [LLM to calculate based on activities_summaries.steps and goals.steps]


## Nutrition

- Total Calories Consumed: [LLM to fill from calories_series.value]
- Macronutrient Breakdown: (percentages or grams)
- Carbohydrates: [LLM to calculate based on food data (assumption needed)]
- Protein: [LLM to calculate based on food data (assumption needed)]
- Fat: [LLM to calculate based on food data (assumption needed)]

### Hydration
- Total Water Intake: [LLM to fill from calories_series.value where activity_description.name = 'Water']

## Body Composition

- Body Weight: [LLM to fill from body_weight_series]
- BMI: [LLM to fill from bmi_series]
- Body Fat Percentage: [LLM to fill from body_fat_series]

## Recovery

- Resting Heart Rate: [LLM to fill from heart_rate_activities.resting_heart_rate]
- HRV: [LLM to fill from heart_rate_variability_time_series]

### Sleep Quality (for reference)

- Reference the previously mentioned Sleep Quality metric from the Sleep Section.

## Additional Information

- Skin Temperature: [LLM to fill from skin_temperatures]
- Core Temperature: [LLM to fill from core_temperatures]
- Oxygen Saturation: [LLM to fill from oxygen_saturation]
- Breathing Rate: [LLM to fill from breathing_rate_series / breathing_rate_intraday]

## Notes

[LLM to add any relevant notes or insights]

## Conclusion

[LLM to add any conclusion about the day's data]