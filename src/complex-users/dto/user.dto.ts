import { ObjectType, Field, InputType, ArgsType } from '@nestjs/graphql';
import {
  IsBoolean,
  IsOptional,
  IsString,
  IsEmail,
  ValidateNested,
  ArrayMaxSize,
  ArrayNotEmpty,
} from 'class-validator';
import { Type } from 'class-transformer';

// Query Arguments
@ArgsType()
export class GetUserPreferencesArgs {
  @Field(() => Boolean, { nullable: true, defaultValue: true })
  @IsOptional()
  @IsBoolean()
  include_channels?: boolean = true;
}

// Input DTOs for Creating/Updating User Preferences
@InputType()
export class QuietHoursInput {
  @Field()
  @IsBoolean()
  enabled: boolean;

  @Field({ nullable: true })
  @IsOptional()
  @IsString()
  start?: string;

  @Field({ nullable: true })
  @IsOptional()
  @IsString()
  end?: string;

  @Field({ nullable: true })
  @IsOptional()
  @IsString()
  timezone?: string;
}

@InputType()
export class DeviceInput {
  @Field()
  @IsString()
  device_id: string;

  @Field()
  @IsString()
  platform: string;

  @Field()
  @IsString()
  token: string;

  @Field()
  @IsBoolean()
  active: boolean;
}

@InputType()
export class EmailChannelInput {
  @Field()
  @IsBoolean()
  enabled: boolean;

  @Field()
  @IsBoolean()
  verified: boolean;

  @Field()
  @IsString()
  frequency: string;

  @Field(() => QuietHoursInput)
  @ValidateNested()
  @Type(() => QuietHoursInput)
  quiet_hours: QuietHoursInput;
}

@InputType()
export class PushChannelInput {
  @Field()
  @IsBoolean()
  enabled: boolean;

  @Field(() => [DeviceInput], { nullable: true })
  @IsOptional()
  @ValidateNested({ each: true })
  @Type(() => DeviceInput)
  devices?: DeviceInput[];

  @Field(() => QuietHoursInput)
  @ValidateNested()
  @Type(() => QuietHoursInput)
  quiet_hours: QuietHoursInput;
}

@InputType()
export class ChannelsInput {
  @Field(() => EmailChannelInput)
  @ValidateNested()
  @Type(() => EmailChannelInput)
  email: EmailChannelInput;

  @Field(() => PushChannelInput)
  @ValidateNested()
  @Type(() => PushChannelInput)
  push: PushChannelInput;
}

@InputType()
export class DigestPreferenceInput {
  @Field()
  @IsBoolean()
  enabled: boolean;

  @Field()
  @IsString()
  frequency: string;

  @Field()
  @IsString()
  time: string;
}

@InputType()
export class UserPreferencesInput {
  @Field()
  @IsBoolean()
  marketing: boolean;

  @Field()
  @IsBoolean()
  transactional: boolean;

  @Field()
  @IsBoolean()
  reminders: boolean;

  @Field(() => DigestPreferenceInput)
  @ValidateNested()
  @Type(() => DigestPreferenceInput)
  digest: DigestPreferenceInput;
}

@InputType()
export class CreateUserPreferencesInput {
  @Field()
  @IsString()
  user_id: string;

  @Field()
  @IsEmail()
  email: string;

  @Field({ nullable: true })
  @IsOptional()
  @IsString()
  phone?: string;

  @Field()
  @IsString()
  timezone: string;

  @Field()
  @IsString()
  language: string;

  @Field()
  @IsBoolean()
  notification_enabled: boolean;

  @Field(() => ChannelsInput, { nullable: true })
  @IsOptional()
  @ValidateNested()
  @Type(() => ChannelsInput)
  channels?: ChannelsInput;

  @Field(() => UserPreferencesInput)
  @ValidateNested()
  @Type(() => UserPreferencesInput)
  preferences: UserPreferencesInput;
}

// Nested DTOs for Response
@ObjectType()
export class QuietHoursDto {
  @Field()
  enabled: boolean;

  @Field({ nullable: true })
  start?: string;

  @Field({ nullable: true })
  end?: string;

  @Field({ nullable: true })
  timezone?: string;
}

@ObjectType()
export class DeviceDto {
  @Field()
  device_id: string;

  @Field()
  platform: string;

  @Field()
  token: string;

  @Field()
  last_seen: Date;

  @Field()
  active: boolean;
}

@ObjectType()
export class EmailChannelDto {
  @Field()
  enabled: boolean;

  @Field()
  verified: boolean;

  @Field()
  frequency: string;

  @Field(() => QuietHoursDto)
  quiet_hours: QuietHoursDto;
}

@ObjectType()
export class PushChannelDto {
  @Field()
  enabled: boolean;

  @Field(() => [DeviceDto])
  devices: DeviceDto[];

  @Field(() => QuietHoursDto)
  quiet_hours: QuietHoursDto;
}

@ObjectType()
export class ChannelsDto {
  @Field(() => EmailChannelDto)
  email: EmailChannelDto;

  @Field(() => PushChannelDto)
  push: PushChannelDto;
}

@ObjectType()
export class DigestPreferenceDto {
  @Field()
  enabled: boolean;

  @Field()
  frequency: string;

  @Field()
  time: string;
}

@ObjectType()
export class UserPreferencesDto {
  @Field()
  marketing: boolean;

  @Field()
  transactional: boolean;

  @Field()
  reminders: boolean;

  @Field(() => DigestPreferenceDto)
  digest: DigestPreferenceDto;
}

// Main Response DTO
@ObjectType()
export class UserPreferencesResponse {
  @Field()
  user_id: string;

  @Field()
  email: string;

  @Field({ nullable: true })
  phone?: string;

  @Field()
  timezone: string;

  @Field()
  language: string;

  @Field()
  notification_enabled: boolean;

  @Field(() => ChannelsDto, { nullable: true })
  channels?: ChannelsDto;

  @Field(() => UserPreferencesDto)
  preferences: UserPreferencesDto;

  @Field()
  updated_at: Date;
}

// Error DTOs
@ObjectType()
export class ErrorDetails {
  @Field({ nullable: true })
  user_id?: string;

  @Field({ nullable: true })
  retry_after?: number;
}

@ObjectType()
export class ErrorDto {
  @Field()
  code: string;

  @Field()
  message: string;

  @Field(() => ErrorDetails)
  details: ErrorDetails;

  @Field()
  request_id: string;
}

@ObjectType()
export class ErrorResponse {
  @Field(() => ErrorDto)
  error: ErrorDto;
}

// Batch Get User Preferences DTOs
@InputType()
export class BatchGetUserPreferencesInput {
  @Field(() => [String])
  @ArrayNotEmpty()
  @ArrayMaxSize(100)
  @IsString({ each: true })
  user_ids: string[];

  @Field(() => Boolean, { nullable: true, defaultValue: true })
  @IsOptional()
  @IsBoolean()
  include_channels?: boolean = true;
}

@ObjectType()
export class BatchGetUserPreferencesResponse {
  @Field(() => [UserPreferencesResponse])
  users: UserPreferencesResponse[];

  @Field(() => [String])
  not_found: string[];

  @Field()
  total_requested: number;

  @Field()
  total_found: number;
}

// Check User Opt-Out Status DTOs
@ObjectType()
export class OptOutChannelsDto {
  @Field()
  email: boolean;

  @Field()
  push: boolean;
}

@ObjectType()
export class OptOutStatusResponse {
  @Field()
  user_id: string;

  @Field()
  opted_out: boolean;

  @Field(() => OptOutChannelsDto)
  channels: OptOutChannelsDto;

  @Field()
  checked_at: Date;
}

// Update Last Notification Time DTOs
@InputType()
export class UpdateLastNotificationInput {
  @Field()
  @IsString()
  channel: string;

  @Field()
  @IsString()
  notification_type: string;

  @Field()
  @IsString()
  notification_id: string;

  @Field()
  @IsString()
  sent_at: string;
}

// Simple User Creation DTOs
export class SimpleUserPreferenceInput {
  @IsBoolean()
  email: boolean;

  @IsBoolean()
  push: boolean;
}

export class CreateSimpleUserInput {
  @IsString()
  name: string;

  @IsEmail()
  email: string;

  @IsString()
  password: string;

  @IsOptional()
  @IsString()
  push_token?: string;

  @ValidateNested()
  @Type(() => SimpleUserPreferenceInput)
  preferences: SimpleUserPreferenceInput;
}

@ObjectType()
export class SimpleUserPreferenceResponse {
  @Field()
  email: boolean;

  @Field()
  push: boolean;
}

@ObjectType()
export class SimpleUserResponse {
  @Field()
  user_id: string;

  @Field()
  name: string;

  @Field()
  email: string;

  @Field({ nullable: true })
  push_token?: string;

  @Field(() => SimpleUserPreferenceResponse)
  preferences: SimpleUserPreferenceResponse;

  @Field()
  created_at: Date;
}
