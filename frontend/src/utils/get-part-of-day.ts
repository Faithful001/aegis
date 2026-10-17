export const getPartOfDay = () => {
  const date = new Date();
  const hour = date.getHours();
  if (hour < 12) {
    return "Morning";
  } else if (hour < 18) {
    return "Afternoon";
  } else {
    return "Evening";
  }
};
