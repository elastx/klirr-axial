import { Avatar } from "@mantine/core";
import { AvatarProps } from "./UserAvatar";


const GroupAvatar: React.FC<AvatarProps> = ({
  seed,
  size = 40,
}) => {
  return (
    <Avatar
      size={size}
      name={seed}
      variant="beam"
    />
  );
}

export default GroupAvatar;