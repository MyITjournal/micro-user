import { Entity, Column, PrimaryColumn, ManyToOne, JoinColumn, CreateDateColumn, UpdateDateColumn } from 'typeorm';
import { ObjectType, Field, ID } from '@nestjs/graphql';
import { UserChannel } from './usersChannel.entity';

@ObjectType()
@Entity('user_devices')
export class UserDevice {
  @Field(() => ID)
  @PrimaryColumn({ type: 'varchar', length: 50 })
  device_id: string;

  @Field()
  @Column({ type: 'uuid' })
  channel_id: string;

  @Field()
  @Column({ type: 'varchar', length: 20 })
  platform: string; // 'ios', 'android', 'web'

  @Field()
  @Column({ type: 'text' })
  token: string;

  @Field()
  @Column({ type: 'timestamp', nullable: true })
  last_seen: Date;

  @Field()
  @Column({ type: 'boolean', default: true })
  active: boolean;

  @Field(() => UserChannel)
  @ManyToOne(() => UserChannel, channel => channel.devices)
  @JoinColumn({ name: 'channel_id' })
  channel: UserChannel;

  @Field()
  @CreateDateColumn()
  created_at: Date;

  @Field()
  @UpdateDateColumn()
  updated_at: Date;
}
