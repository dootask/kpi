"use client"

import * as React from "react"
import { ChevronDownIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"

interface DateTimePickerProps {
  value?: string
  onChange?: (value: string) => void
  label?: string
  id?: string
  className?: string
}

export function DateTimePicker({ value, onChange, label, id, className }: DateTimePickerProps) {
  const [open, setOpen] = React.useState(false)
  const [date, setDate] = React.useState(value ? value.split('T')[0] : '')
  const [time, setTime] = React.useState(value ? value.split('T')[1] : '09:00')

  React.useEffect(() => {
    if (value) {
      const [datePart, timePart] = value.split('T')
      setDate(datePart || '')
      setTime(timePart || '09:00')
    }
  }, [value])

  const handleDateChange = (newDate: string) => {
    setDate(newDate)
    if (newDate && time) {
      onChange?.(`${newDate}T${time}`)
    }
  }

  const handleTimeChange = (newTime: string) => {
    setTime(newTime)
    if (date && newTime) {
      onChange?.(`${date}T${newTime}`)
    }
  }

  const displayValue = React.useMemo(() => {
    if (!date) return "选择日期和时间"
    const dateObj = new Date(`${date}T${time}`)
    return dateObj.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })
  }, [date, time])

  return (
    <div className={`flex flex-col gap-2 ${className}`}>
      {label && (
        <Label htmlFor={id} className="px-1">
          {label}
        </Label>
      )}
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            id={id}
            className="w-full justify-between font-normal"
          >
            {displayValue}
            <ChevronDownIcon className="w-4 h-4" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto overflow-hidden p-4" align="start">
          <div className="flex gap-4">
            <div className="flex flex-col gap-3">
              <Label htmlFor="date-picker" className="px-1">
                日期
              </Label>
              <Input
                type="date"
                id="date-picker"
                value={date}
                onChange={(e) => handleDateChange(e.target.value)}
                className="w-40"
              />
            </div>
            <div className="flex flex-col gap-3">
              <Label htmlFor="time-picker" className="px-1">
                时间
              </Label>
              <Input
                type="time"
                id="time-picker"
                value={time}
                onChange={(e) => handleTimeChange(e.target.value)}
                className="w-32"
              />
            </div>
          </div>
          <div className="flex justify-end mt-4">
            <Button size="sm" onClick={() => setOpen(false)}>
              确定
            </Button>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  )
}